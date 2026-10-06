package routes

import (
	"encoding/json"
	"net/http"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase/pluginsystem"
)

type pluginSessionAuthValidateRequest struct {
	PluginID    string          `json:"pluginId"`
	InstanceID  string          `json:"instanceId,omitempty"`
	AuthContext string          `json:"authContext,omitempty"`
	Auth        map[string]any  `json:"auth,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
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
	auth := map[string]any{}
	config := installedPluginConfig(e.App, data.PluginID)
	plaintext := map[string]bool{}
	instanceID := data.InstanceID
	if instanceID != "" {
		instance, err := pluginAuthInstance(e.App, e.Auth.Id, data.PluginID, instanceID)
		if err != nil {
			return err
		}
		instanceID = instance.Id
		auth = pluginsystem.JSONMapFromRecord(instance, "auth")
		pluginsystem.MergePluginConfig(config, pluginsystem.JSONMapFromRecord(instance, "config"))
	}
	requested := data.AuthContext
	if requested == "" {
		resolved, _, resolveErr := pluginsystem.EffectiveAuthContext(plugin.Manifest, auth)
		if resolveErr != nil {
			return pluginOperationAPIError(pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
		}
		requested = resolved
	}
	contextName, authContext, err := sessionAuthContext(plugin, requested)
	if err != nil {
		return pluginOperationAPIError(pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
	}
	if authContext.Refresh == nil || authContext.Refresh.Function == "" {
		return pluginOperationAPIError(pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
	}

	for key, value := range data.Auth {
		if value != nil {
			if _, ok := value.(string); !ok {
				return apis.NewBadRequestError("auth credentials must be strings", nil)
			}
		}
		if value == "" {
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
			return pluginOperationAPIError(pluginsystem.Failure("invalid_configuration", pluginsystem.OperationLogin))
		}
	}
	pluginsystem.MergePluginConfig(config, draft)
	policy, err := pluginInstancePolicy(plugin, config)
	if err != nil {
		pluginsystem.LogPluginFailure(e.App.Logger(), "plugin configuration rejected", err, "plugin", plugin.Manifest.ID)
		return pluginOperationAPIError(err)
	}
	auth = pluginsystem.AuthForPluginRefresh(auth, authContext)
	if err := decryptActiveAuth(auth, authContext, plaintext); err != nil {
		return pluginOperationAPIError(pluginAuthFailure(err, pluginsystem.OperationLogin))
	}

	for _, field := range authContext.Fields {
		if pluginsystem.StringFromAny(auth[field]) == "" {
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
		return err
	}

	runtime, err := runtimeFor(plugin)
	if err != nil {
		return err
	}
	policy = policy.WithHostAuth(auth)
	output, err := runtime.Call(e.Request.Context(), plugin, authContext.Refresh.Function, inputBytes, policy)
	if err != nil {
		return pluginOperationAPIError(err)
	}
	if err := pluginsystem.ValidatePluginSessionRefreshOutput(output); err != nil {
		return pluginOperationAPIError(err)
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
