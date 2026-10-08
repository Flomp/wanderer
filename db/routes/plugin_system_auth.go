package routes

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase/pluginsystem"
)

type pluginOAuthStartRequest struct {
	PluginID           string `json:"pluginId"`
	InstanceID         string `json:"instanceId"`
	AuthContext        string `json:"authContext,omitempty"`
	RedirectURI        string `json:"redirectUri"`
	ExpectedGeneration string `json:"expected_connection_generation"`
	ActivateOnSuccess  bool   `json:"activate_on_success"`
}

type pluginOAuthCallbackRequest struct {
	InstanceID string `json:"instanceId"`
	Code       string `json:"code"`
	State      string `json:"state"`
}

type pluginOAuthRevokeRequest struct {
	InstanceID         string `json:"instanceId"`
	ExpectedGeneration string `json:"expected_connection_generation"`
}

func PluginSystemOAuthStart(e *core.RequestEvent) error {
	logContext := pluginOperationLogContext{Action: pluginActionOAuthStart}
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}

	var data pluginOAuthStartRequest
	if err := e.BindBody(&data); err != nil {
		return apis.NewBadRequestError("failed to read request data", nil)
	}
	if data.PluginID == "" || data.RedirectURI == "" {
		return apis.NewBadRequestError("pluginId and redirectUri are required", nil)
	}
	if err := pluginsystem.ValidateOAuthRedirectURI(data.RedirectURI); err != nil {
		return apis.NewBadRequestError("redirectUri is not allowed", nil)
	}

	plugin, err := localPlugin(e.App, data.PluginID)
	if err != nil {
		return err
	}
	logContext.PluginID = plugin.Manifest.ID
	instance, err := pluginAuthInstance(e.App, e.Auth.Id, data.PluginID, data.InstanceID)
	if err != nil {
		return err
	}
	logContext.InstanceID = instance.Id
	if data.ExpectedGeneration == "" {
		return pluginsystem.ConnectionGenerationRequiredError()
	}
	snapshot := pluginsystem.CaptureConnection(instance)
	if snapshot.Generation != data.ExpectedGeneration {
		return pluginsystem.ConnectionAPIError(pluginsystem.ErrConnectionConflict)
	}
	requested := data.AuthContext
	if requested == "" {
		if _, _, err := pluginsystem.EffectiveAuthContext(plugin.Manifest, instance.GetString("auth_context")); err != nil {
			return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.Failure("invalid_configuration", pluginsystem.OperationOAuthExchange))
		}
		requested = instance.GetString("auth_context")
	}
	contextName, authContext, err := pluginsystem.OAuthContext(plugin, requested)
	if err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.Failure("invalid_configuration", pluginsystem.OperationOAuthExchange))
	}
	auth, err := decryptedOAuthFlowAuth(instance, plugin.Manifest, authContext, false)
	if err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginAuthFailure(err, pluginsystem.OperationOAuthExchange))
	}
	clientID := pluginsystem.StringFromAny(auth["clientId"])
	if clientID == "" {
		logPluginOperationFailure(e, logContext, pluginsystem.WithDiagnosticCause(pluginsystem.Failure("invalid_request", pluginsystem.OperationOAuthExchange), "host_auth_invalid"))
		return apis.NewBadRequestError("oauth clientId is required", nil)
	}

	state := pluginsystem.NewOAuthState(32)
	patch := map[string]any{pluginsystem.AuthFieldOAuthFlowContext: contextName, pluginsystem.AuthFieldOAuthActivateOnCallback: data.ActivateOnSuccess, pluginsystem.AuthFieldOAuthState: state, pluginsystem.AuthFieldOAuthRedirectURI: data.RedirectURI}

	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", clientID)
	values.Set("redirect_uri", data.RedirectURI)
	values.Set("state", state)
	if len(authContext.Scopes) > 0 {
		separator := authContext.ScopeSeparator
		if separator == "" {
			separator = " "
		}
		values.Set("scope", strings.Join(authContext.Scopes, separator))
	}
	for key, value := range authContext.AuthorizationParams {
		values.Set(key, value)
	}
	// A fresh non-PKCE flow must discard a stale verifier without reading it.
	patch[pluginsystem.AuthFieldOAuthCodeVerifier] = nil
	if authContext.PKCE {
		verifier := pluginsystem.NewOAuthCodeVerifier(64)
		patch[pluginsystem.AuthFieldOAuthCodeVerifier] = verifier
		values.Set("code_challenge_method", "S256")
		values.Set("code_challenge", pluginsystem.PKCEChallenge(verifier))
	}

	committed, err := pluginsystem.PatchConnection(e.App, snapshot, false, func(fresh *core.Record) error {
		return pluginsystem.ApplyAuthPatch(fresh, plugin.Manifest, patch)
	})
	if err != nil {
		if errors.Is(err, pluginsystem.ErrConnectionObsolete) {
			return pluginConnectionAPIErrorForEvent(e, logContext, err)
		}
		return pluginHostAPIErrorForEvent(e, logContext, "Failed to save plugin connection.", pluginsystem.WithDiagnosticCause(err, "database_save"))
	}

	authURL, err := url.Parse(authContext.AuthorizationURL)
	if err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.WithDiagnosticCause(pluginsystem.Failure("invalid_configuration", pluginsystem.OperationOAuthExchange), "request_build_failed"))
	}
	query := authURL.Query()
	for key, value := range values {
		query[key] = value
	}
	authURL.RawQuery = query.Encode()

	return e.JSON(http.StatusOK, map[string]any{
		"url":                   authURL.String(),
		"state":                 state,
		"instanceId":            instance.Id,
		"connection_generation": committed.GetString("connection_generation"),
		"updated":               committed.GetString("updated"),
	})
}

func PluginSystemOAuthCallback(e *core.RequestEvent) error {
	logContext := pluginOperationLogContext{Action: pluginActionOAuthCallback}
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}

	var data pluginOAuthCallbackRequest
	if err := e.BindBody(&data); err != nil {
		return apis.NewBadRequestError("failed to read request data", nil)
	}
	if data.InstanceID == "" || data.Code == "" || data.State == "" {
		return apis.NewBadRequestError("instanceId, code and state are required", nil)
	}

	instance, err := e.App.FindRecordById("plugin_instances", data.InstanceID)
	if err != nil || instance.GetString("user") != e.Auth.Id {
		return apis.NewNotFoundError("plugin instance not found", nil)
	}
	logContext.InstanceID = instance.Id
	plugin, err := localPlugin(e.App, instance.GetString("plugin_id"))
	if err != nil {
		return err
	}
	logContext.PluginID = plugin.Manifest.ID
	snapshot := pluginsystem.CaptureConnection(instance)
	instance, err = pluginsystem.CheckConnection(e.App, snapshot, false)
	if err != nil {
		return pluginConnectionAPIErrorForEvent(e, logContext, err)
	}
	storedAuth := pluginsystem.JSONMapFromRecord(instance, "auth")
	contextName := pluginsystem.StringFromAny(storedAuth[pluginsystem.AuthFieldOAuthFlowContext])
	if contextName == "" {
		return pluginsystem.ConnectionAPIError(pluginsystem.ErrOAuthFlowConflict)
	}
	_, authContext, err := pluginsystem.OAuthContext(plugin, contextName)
	if err != nil || authContext.Type != pluginsystem.AuthTypeOAuth2 {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.Failure("invalid_configuration", pluginsystem.OperationOAuthExchange))
	}
	auth, err := decryptedOAuthFlowAuth(instance, plugin.Manifest, authContext, true)
	if err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginAuthFailure(err, pluginsystem.OperationOAuthExchange))
	}
	if data.State != pluginsystem.StringFromAny(auth[pluginsystem.AuthFieldOAuthState]) {
		logPluginOperationFailure(e, logContext, pluginsystem.WithDiagnosticCause(pluginsystem.Failure("invalid_request", pluginsystem.OperationOAuthExchange), "host_auth_invalid"))
		return pluginsystem.ConnectionAPIError(pluginsystem.ErrOAuthFlowConflict)
	}

	values := map[string]string{
		"grant_type":   "authorization_code",
		"code":         data.Code,
		"redirect_uri": pluginsystem.StringFromAny(auth[pluginsystem.AuthFieldOAuthRedirectURI]),
	}
	if authContext.PKCE {
		values["code_verifier"] = pluginsystem.StringFromAny(auth[pluginsystem.AuthFieldOAuthCodeVerifier])
	}
	token, err := pluginsystem.ExchangeOAuthToken(e.Request.Context(), plugin.Manifest, authContext, auth, values)
	if err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, err)
	}
	patch := map[string]any{}
	pluginsystem.ClearOAuthToken(patch)
	pluginsystem.StoreOAuthToken(patch, contextName, token)
	for _, field := range pluginsystem.InternalOAuthTransientFields() {
		patch[field] = nil
	}
	_, err = pluginsystem.PatchConnection(e.App, snapshot, false, func(fresh *core.Record) error {
		flow, decryptErr := decryptedOAuthFlowAuth(fresh, plugin.Manifest, authContext, true)
		if decryptErr != nil {
			return decryptErr
		}
		if pluginsystem.StringFromAny(flow[pluginsystem.AuthFieldOAuthState]) != data.State || pluginsystem.StringFromAny(flow[pluginsystem.AuthFieldOAuthFlowContext]) != contextName {
			return pluginsystem.ErrOAuthFlowConflict
		}
		activate, _ := flow[pluginsystem.AuthFieldOAuthActivateOnCallback].(bool)
		if err := pluginsystem.ApplyAuthPatch(fresh, plugin.Manifest, patch); err != nil {
			return err
		}
		fresh.Set("auth_context", contextName)
		fresh.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
		pluginsystem.ResetConnection(fresh)
		if activate {
			fresh.Set("enabled", true)
		}
		if fresh.GetBool("enabled") {
			fresh.Set("status", "configured")
		} else {
			fresh.Set("status", "disabled")
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, pluginsystem.ErrConnectionObsolete) || errors.Is(err, pluginsystem.ErrOAuthFlowConflict) {
			return pluginConnectionAPIErrorForEvent(e, logContext, err)
		}
		return pluginHostAPIErrorForEvent(e, logContext, "Failed to save plugin connection.", pluginsystem.WithDiagnosticCause(err, "database_save"))
	}

	return e.JSON(http.StatusOK, map[string]any{"ok": true})
}

func PluginSystemOAuthRevoke(e *core.RequestEvent) error {
	logContext := pluginOperationLogContext{Action: pluginActionOAuthRevoke}
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}

	var data pluginOAuthRevokeRequest
	if err := e.BindBody(&data); err != nil {
		return apis.NewBadRequestError("failed to read request data", nil)
	}
	if data.InstanceID == "" {
		return apis.NewBadRequestError("instanceId is required", nil)
	}
	instance, err := e.App.FindRecordById("plugin_instances", data.InstanceID)
	if err != nil || instance.GetString("user") != e.Auth.Id {
		return apis.NewNotFoundError("plugin instance not found", nil)
	}
	logContext.InstanceID = instance.Id
	if data.ExpectedGeneration == "" {
		return pluginsystem.ConnectionGenerationRequiredError()
	}
	snapshot := pluginsystem.CaptureConnection(instance)
	if snapshot.Generation != data.ExpectedGeneration {
		return pluginsystem.ConnectionAPIError(pluginsystem.ErrConnectionConflict)
	}
	committed, err := pluginsystem.PatchConnection(e.App, snapshot, false, func(fresh *core.Record) error {
		auth := pluginsystem.JSONMapFromRecord(fresh, "auth")
		pluginsystem.ClearOAuthToken(auth)
		// The revoke patch contains removals only: no old secret is decrypted or
		// rewritten, including grants encrypted with a missing key.
		patch := map[string]any{}
		for field, value := range auth {
			if value == nil {
				patch[field] = nil
			}
		}
		if err := pluginsystem.ApplyAuthPatch(fresh, pluginsystem.Manifest{}, patch); err != nil {
			return err
		}
		fresh.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
		pluginsystem.ResetConnection(fresh)
		fresh.Set("enabled", false)
		fresh.Set("status", "disabled")
		return nil
	})
	if err != nil {
		if errors.Is(err, pluginsystem.ErrConnectionObsolete) {
			return pluginConnectionAPIErrorForEvent(e, logContext, err)
		}
		return pluginHostAPIErrorForEvent(e, logContext, "Failed to save plugin connection.", pluginsystem.WithDiagnosticCause(err, "database_save"))
	}

	return e.JSON(http.StatusOK, map[string]any{"ok": true, "connection_generation": committed.GetString("connection_generation"), "updated": committed.GetString("updated")})
}

func pluginAuthInstance(app core.App, userID string, pluginID string, instanceID string) (*core.Record, error) {
	if instanceID != "" {
		instance, err := app.FindRecordById("plugin_instances", instanceID)
		if err != nil || instance.GetString("user") != userID || instance.GetString("plugin_id") != pluginID {
			return nil, apis.NewNotFoundError("plugin instance not found", nil)
		}
		return instance, nil
	}
	instance, err := app.FindFirstRecordByFilter(
		"plugin_instances",
		"user={:user} && plugin_id={:plugin_id}",
		dbx.Params{"user": userID, "plugin_id": pluginID},
	)
	if err != nil {
		return nil, apis.NewNotFoundError("plugin instance not found", nil)
	}
	return instance, nil
}
