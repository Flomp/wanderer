package routes

import (
	"encoding/json"
	"net/http"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase/pluginsystem"
)

type pluginSessionAuthValidateRequest struct {
	PluginID           string          `json:"pluginId"`
	InstanceID         string          `json:"instanceId,omitempty"`
	AuthContext        string          `json:"authContext,omitempty"`
	Auth               map[string]any  `json:"auth,omitempty"`
	Config             json.RawMessage `json:"config,omitempty"`
	ExpectedGeneration string          `json:"expected_connection_generation"`
}

type pluginSessionAuthRefreshInput struct {
	Instance pluginsystem.InstanceRef `json:"instance"`
	Auth     map[string]any           `json:"auth,omitempty"`
	Config   map[string]any           `json:"config,omitempty"`
}

func PluginSystemSessionAuthValidate(e *core.RequestEvent) error {
	return pluginSystemSessionAuthValidate(e, pluginsystem.NewRuntimeRegistry().RuntimeFor)
}

func pluginSystemSessionAuthValidate(e *core.RequestEvent, runtimeFor func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error)) error {
	logContext := pluginOperationLogContext{Action: pluginActionValidate}
	if e.Auth == nil {
		return apis.NewUnauthorizedError("authentication required", nil)
	}

	var data pluginSessionAuthValidateRequest
	if err := e.BindBody(&data); err != nil {
		return apis.NewBadRequestError("failed to read request data", nil)
	}
	if data.PluginID == "" {
		return apis.NewBadRequestError("pluginId is required", nil)
	}

	plugin, err := localPlugin(e.App, data.PluginID)
	if err != nil {
		return err
	}
	logContext.PluginID = plugin.Manifest.ID
	auth := map[string]any{}
	config := installedPluginConfig(e.App, data.PluginID)
	plaintext := map[string]bool{}
	instanceID := data.InstanceID
	var snapshot pluginsystem.ConnectionSnapshot
	selected := ""
	if instanceID != "" {
		instance, err := pluginAuthInstance(e.App, e.Auth.Id, data.PluginID, instanceID)
		if err != nil {
			return err
		}
		instanceID = instance.Id
		logContext.InstanceID = instance.Id
		if data.ExpectedGeneration == "" {
			return pluginsystem.ConnectionGenerationRequiredError()
		}
		snapshot = pluginsystem.CaptureConnection(instance)
		if snapshot.Generation != data.ExpectedGeneration {
			return pluginsystem.ConnectionAPIError(pluginsystem.ErrConnectionConflict)
		}
		release, acquired := pluginsystem.TryLockInstance(instance.Id)
		if !acquired {
			return pluginsystem.ConnectionAPIError(pluginsystem.ErrInstanceBusy)
		}
		defer release()
		instance, err = pluginsystem.CheckConnection(e.App, snapshot, false)
		if err != nil {
			return pluginConnectionAPIErrorForEvent(e, logContext, err)
		}
		selected = instance.GetString("auth_context")
		auth = pluginsystem.JSONMapFromRecord(instance, "auth")
		pluginsystem.MergePluginConfig(config, pluginsystem.JSONMapFromRecord(instance, "config"))
	}
	requested := data.AuthContext
	if requested == "" {
		var resolved string
		var resolveErr error
		if instanceID == "" {
			resolved, _, resolveErr = pluginsystem.ResolveAuthContext(plugin.Manifest, "")
		} else {
			resolved, _, resolveErr = pluginsystem.EffectiveAuthContext(plugin.Manifest, selected)
		}
		if resolveErr != nil {
			return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
		}
		requested = resolved
	}
	contextName, authContext, err := sessionAuthContext(plugin, requested)
	if err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
	}
	if authContext.Refresh == nil || authContext.Refresh.Function == "" {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
	}

	editable := pluginsystem.EditableAuthCredentialFields(plugin.Manifest, contextName)
	secrets := pluginsystem.ManifestAuthSecretFields(plugin.Manifest)
	for key, value := range data.Auth {
		if !editable[key] {
			return apis.NewBadRequestError("Unsupported credential field.", nil)
		}
		if value != nil {
			if _, ok := value.(string); !ok {
				return apis.NewBadRequestError("auth credentials must be strings", nil)
			}
		}
		if value == nil {
			if !secrets[key] {
				return apis.NewBadRequestError("Only editable secret credentials can be removed with null.", nil)
			}
			delete(auth, key)
			continue
		}
		if value == "" && secrets[key] {
			continue
		}
		auth[key] = value
		plaintext[key] = true
	}
	// Apply fresh stored overrides then the unsaved candidate, with Settings merge
	// semantics, before resolving policy or running any provider code.
	var draft map[string]any
	if len(data.Config) > 0 {
		if err := json.Unmarshal(data.Config, &draft); err != nil || draft == nil {
			return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
		}
	}
	pluginsystem.MergePluginConfig(config, draft)
	policy, err := pluginInstancePolicy(plugin, config)
	if err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, err)
	}
	auth = pluginsystem.AuthForPluginRefresh(auth, authContext)
	if err := decryptActiveAuth(auth, plugin.Manifest, authContext, plaintext); err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginAuthFailure(err, pluginsystem.OperationLogin))
	}

	for _, field := range append(append([]string{}, authContext.Fields...), authContext.SecretFields...) {
		if pluginsystem.StringFromAny(auth[field]) == "" {
			logPluginOperationFailure(e, logContext, pluginsystem.WithDiagnosticCause(pluginsystem.Failure("invalid_request", pluginsystem.OperationLogin), "host_auth_invalid"))
			return apis.NewBadRequestError("required session credentials are missing", nil)
		}
	}
	inputBytes, err := json.Marshal(pluginSessionAuthRefreshInput{
		Instance: pluginsystem.InstanceRef{
			ID:       instanceID,
			PluginID: plugin.Manifest.ID,
		},
		Auth:   auth,
		Config: pluginRuntimeConfig(config),
	})
	if err != nil {
		return pluginHostAPIErrorForEvent(e, logContext, "Failed to prepare plugin request.", pluginsystem.WithDiagnosticCause(err, "input_encoding"))
	}

	runtime, err := runtimeFor(plugin)
	if err != nil {
		return pluginHostAPIErrorForEvent(e, logContext, "Failed to initialize plugin runtime.", err)
	}
	policy = policy.WithHostAuth(auth, contextName)
	if instanceID != "" {
		if _, err := pluginsystem.CheckConnection(e.App, snapshot, false); err != nil {
			return pluginConnectionAPIErrorForEvent(e, logContext, err)
		}
	}
	output, err := runtime.Call(e.Request.Context(), plugin, authContext.Refresh.Function, inputBytes, policy)
	if instanceID != "" {
		if _, guardErr := pluginsystem.CheckConnection(e.App, snapshot, false); guardErr != nil {
			return pluginConnectionAPIErrorForEvent(e, logContext, guardErr)
		}
	}
	if err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, err)
	}
	if err := pluginsystem.ValidatePluginSessionRefreshOutput(output); err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, err)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"ok":          true,
		"authContext": contextName,
	})
}

func sessionAuthContext(plugin pluginsystem.LocalPlugin, requested string) (string, pluginsystem.AuthContext, error) {
	name, auth, err := pluginsystem.ResolveAuthContext(plugin.Manifest, requested)
	if err != nil {
		return "", pluginsystem.AuthContext{}, err
	}
	if auth.Type != pluginsystem.AuthTypeSession {
		return "", pluginsystem.AuthContext{}, apis.NewBadRequestError("session auth context not found", nil)
	}
	return name, auth, nil
}
