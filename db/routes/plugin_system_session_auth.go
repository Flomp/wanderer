package routes

import (
	"encoding/json"
	"net/http"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase/pluginsystem"
)

type pluginSessionAuthValidateRequest struct {
	PluginID    string         `json:"pluginId"`
	InstanceID  string         `json:"instanceId,omitempty"`
	AuthContext string         `json:"authContext,omitempty"`
	Auth        map[string]any `json:"auth,omitempty"`
}

type pluginSessionAuthRefreshInput struct {
	Instance pluginsystem.InstanceRef `json:"instance"`
	Auth     map[string]any           `json:"auth,omitempty"`
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
		return apis.NewBadRequestError("failed to read request data", err)
	}
	if data.PluginID == "" {
		return apis.NewBadRequestError("pluginId is required", nil)
	}

	plugin, err := localPlugin(e.App, data.PluginID)
	if err != nil {
		return err
	}
	logContext.PluginID = plugin.Manifest.ID
	contextName, authContext, err := sessionAuthContext(plugin, data.AuthContext)
	if err != nil {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
	}
	if authContext.Refresh == nil || authContext.Refresh.Function == "" {
		return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
	}

	auth := map[string]any{}
	instanceID := data.InstanceID
	if instanceID != "" {
		instance, err := pluginAuthInstance(e.App, e.Auth.Id, data.PluginID, instanceID)
		if err != nil {
			return err
		}
		instanceID = instance.Id
		logContext.InstanceID = instance.Id
		auth, err = decryptedInstanceAuth(instance)
		if err != nil {
			return pluginOperationAPIErrorForEvent(e, logContext, pluginsystem.WithDiagnosticCause(pluginsystem.Failure("credentials_unreadable", pluginsystem.OperationLogin), "credentials_unreadable"))
		}
	}
	for key, value := range data.Auth {
		if value == "" {
			continue
		}
		auth[key] = value
	}

	inputBytes, err := json.Marshal(pluginSessionAuthRefreshInput{
		Instance: pluginsystem.InstanceRef{
			ID:       instanceID,
			PluginID: plugin.Manifest.ID,
		},
		Auth: pluginsystem.AuthForPluginRefresh(auth, authContext),
	})
	if err != nil {
		logPluginOperationFailure(e, logContext, pluginsystem.WithDiagnosticCause(err, "input_encoding"))
		return err
	}

	runtime, err := runtimeFor(plugin)
	if err != nil {
		logPluginOperationFailure(e, logContext, err)
		return err
	}
	// TODO: accept and merge plugin instance config here before supporting
	// session-auth plugins with configured connectors. The current validation
	// path is sufficient for public_api session plugins such as komoot and
	// hammerhead, but configured connectors need host config for policy
	// resolution and refresh input parity with production auth injection.
	policy := pluginInstancePolicy(plugin, map[string]any{}).WithHostAuth(auth)
	output, err := runtime.Call(e.Request.Context(), plugin, authContext.Refresh.Function, inputBytes, policy)
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
	if requested != "" {
		authContext, ok := plugin.Manifest.Auth.Contexts[requested]
		if !ok || authContext.Type != pluginsystem.AuthTypeSession {
			return "", pluginsystem.AuthContext{}, apis.NewBadRequestError("unknown session auth context", nil)
		}
		return requested, authContext, nil
	}
	for name, authContext := range plugin.Manifest.Auth.Contexts {
		if authContext.Type == pluginsystem.AuthTypeSession {
			return name, authContext, nil
		}
	}
	return "", pluginsystem.AuthContext{}, apis.NewBadRequestError("session auth context not found", nil)
}
