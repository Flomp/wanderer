package routes

import (
	"errors"
	"log/slog"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase/pluginsystem"
)

// PocketBase preserves codes only for SafeErrorItems (validation.Error). Raw
// scalar maps are rewritten to validation_invalid_value. Keep safe metadata in
// params so proxies and the UI can distinguish errors without parsing text.
func pluginOperationAPIError(err error) error {
	safe := pluginsystem.PublicPluginError(err)
	params := map[string]any{}
	if safe.Operation != "" {
		params["operation"] = safe.Operation
	}
	if safe.HTTPStatus != 0 {
		params["httpStatus"] = safe.HTTPStatus
	}
	if safe.RetryAfterSeconds != nil {
		params["retryAfterSeconds"] = *safe.RetryAfterSeconds
	}
	return apis.NewBadRequestError("Plugin operation failed.", validation.Errors{
		"operation": validation.NewError(safe.Code, safe.Message).SetParams(params),
	})
}

func firstProviderHeader(headers map[string][]string, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// Active ciphertext failure requires credential repair. An unknown or ambiguous
// context is a configuration error and must not masquerade as a rejected grant.
func pluginAuthFailure(err error, operation string) error {
	code := "invalid_configuration"
	if errors.Is(err, pluginsystem.ErrActiveCredentialsUnreadable) {
		code = "credentials_unreadable"
	}
	cause := "auth_context_invalid"
	if code == "credentials_unreadable" {
		cause = "credentials_unreadable"
	}
	return pluginsystem.WithDiagnosticCause(pluginsystem.Failure(code, operation), cause)
}

// References are populated only after host lookup/ownership checks. Request
// payload IDs, provider names, URLs, bodies and credentials are not log attrs.
type pluginOperationLogContext struct {
	Action     pluginOperationAction
	PluginID   string
	InstanceID string
}

type pluginOperationAction string

const (
	pluginActionOAuthStart    pluginOperationAction = "oauth_start"
	pluginActionOAuthCallback pluginOperationAction = "oauth_callback"
	pluginActionOAuthRevoke   pluginOperationAction = "oauth_revoke"
	pluginActionSend          pluginOperationAction = "send"
	pluginActionValidate      pluginOperationAction = "session_validate"
)

// Keep the existing API envelope, and log the typed failure before conversion
// discards its cause chain. Interactive paths use the same safe diagnostics as
// cron, without treating the provider error text as an operational diagnostic.
func pluginOperationAPIErrorForEvent(e *core.RequestEvent, context pluginOperationLogContext, err error) error {
	logPluginOperationFailure(e, context, err)
	return pluginOperationAPIError(err)
}

// Raw local errors keep their existing HTTP/API response; call this before
// returning them to record a safe diagnostic without changing that contract.
func logPluginOperationFailure(e *core.RequestEvent, context pluginOperationLogContext, err error) {
	action := "plugin_operation"
	switch context.Action {
	case pluginActionOAuthStart, pluginActionOAuthCallback, pluginActionOAuthRevoke, pluginActionSend, pluginActionValidate:
		action = string(context.Action)
	}
	attrs := []any{"action", action}
	if e.Auth != nil {
		attrs = append(attrs, "owner", e.Auth.Id)
	}
	if context.PluginID != "" {
		attrs = append(attrs, "plugin", context.PluginID)
	}
	if context.InstanceID != "" {
		attrs = append(attrs, "instance", context.InstanceID)
	}
	diagnostic := pluginsystem.PluginDiagnosticAttrs(err)
	level := slog.LevelWarn
	// Untyped local failures have no provider operation. The closed host action
	// identifies their boundary without changing the public error envelope.
	for i := 0; i+1 < len(diagnostic); i += 2 {
		if diagnostic[i] == "operation" && diagnostic[i+1] == "" {
			diagnostic[i+1] = action
		}
		// An absent track is an expected send prerequisite, not a failed
		// filesystem read or provider operation. Keep actual read errors WARN.
		if diagnostic[i] == "cause" && diagnostic[i+1] == "trail_gpx_missing" {
			level = slog.LevelInfo
		}
	}
	e.App.Logger().Log(e.Request.Context(), level, "plugin operation failed", append(attrs, diagnostic...)...)
}
