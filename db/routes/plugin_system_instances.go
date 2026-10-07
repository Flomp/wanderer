package routes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/pocketbase/pocketbase/tools/security"

	"pocketbase/pluginsystem"
	"pocketbase/util"
)

// The optional test barrier pauses before any write transaction is acquired.
var pluginInstanceSettingsBeforeCommit func()

func bindPluginInstanceSettings(e *core.RequestEvent, create bool) (map[string]any, error) {
	var dto map[string]any
	decoder := json.NewDecoder(e.Request.Body)
	if err := decoder.Decode(&dto); err != nil || dto == nil {
		return nil, apis.NewBadRequestError("A JSON object is required.", nil)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, apis.NewBadRequestError("A single JSON object is required.", nil)
	}
	allowed := map[string]bool{"config": true, "enabled": true, "auth": true, "authContext": true}
	if create {
		allowed["plugin_id"] = true
	} else {
		allowed["expected_connection_generation"] = true
	}
	for key, value := range dto {
		if !allowed[key] {
			return nil, apis.NewBadRequestError("Unsupported plugin instance settings field.", nil)
		}
		switch key {
		case "config", "auth":
			if _, ok := value.(map[string]any); !ok {
				return nil, apis.NewBadRequestError("Config and auth must be JSON objects.", nil)
			}
		case "enabled":
			if _, ok := value.(bool); !ok {
				return nil, apis.NewBadRequestError("Enabled must be a boolean.", nil)
			}
		default:
			if _, ok := value.(string); !ok {
				return nil, apis.NewBadRequestError("Plugin and connection selections must be strings.", nil)
			}
		}
	}
	return dto, nil
}

func PluginSystemInstanceCreate(e *core.RequestEvent) error {
	if e.Auth == nil || e.HasSuperuserAuth() {
		return apis.NewForbiddenError("An authenticated plugin owner is required.", nil)
	}
	dto, err := bindPluginInstanceSettings(e, true)
	if err != nil {
		return err
	}
	pluginID, _ := dto["plugin_id"].(string)
	if pluginID == "" {
		return apis.NewBadRequestError("A plugin id is required.", nil)
	}
	var saved *core.Record
	err = e.App.RunInTransaction(func(tx core.App) error {
		manifest, err := pluginsystem.InstalledPluginManifest(tx, pluginID)
		if err != nil {
			return apis.NewBadRequestError("The installed plugin is unavailable.", nil)
		}
		existing, err := tx.FindAllRecords("plugin_instances", dbx.HashExp{"user": e.Auth.Id, "plugin_id": pluginID})
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return pluginsystem.ErrInstanceExists
		}
		collection, err := tx.FindCollectionByNameOrId("plugin_instances")
		if err != nil {
			return err
		}
		record := core.NewRecord(collection)
		record.Set("user", e.Auth.Id)
		record.Set("plugin_id", pluginID)
		record.Set("connection_generation", pluginsystem.NewConnectionRevision())
		record.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
		selected, _ := dto["authContext"].(string)
		if len(manifest.Auth.Contexts) > 0 {
			selected, _, err = pluginsystem.ResolveAuthContext(manifest, selected)
			if err != nil {
				return apis.NewBadRequestError("Select a valid authentication context.", nil)
			}
		} else if selected != "" {
			return apis.NewBadRequestError("This plugin has no authentication context.", nil)
		}
		record.Set("auth_context", selected)
		auth, _, _, err := applyClientCredentialPatch(record, manifest, selected, dto["auth"])
		if err != nil {
			return err
		}
		record.Set("auth", auth)
		config := installedPluginConfig(tx, pluginID)
		if patch, ok := dto["config"].(map[string]any); ok {
			pluginsystem.MergePluginConfig(config, patch)
		}
		if _, err := pluginInstancePolicy(pluginsystem.LocalPlugin{Manifest: manifest}, config); err != nil {
			return pluginOperationAPIError(err)
		}
		record.Set("config", config)
		record.Set("enabled", dto["enabled"] == true)
		status := pluginsystem.LocalConnectionStatus(manifest, selected, auth, record.GetBool("enabled"), time.Now())
		if record.GetBool("enabled") && status != "configured" {
			return apis.NewBadRequestError("The plugin connection is not ready to enable.", nil)
		}
		record.Set("status", status)
		record.Set("state", map[string]any{})
		record.Set("last_error", map[string]any{})
		if err := tx.Save(record); err != nil {
			return err
		}
		saved = record.Fresh()
		return nil
	})
	if err != nil {
		return pluginInstanceWriteError(e, err)
	}
	public, err := pluginsystem.PublicConnectionRecord(e.App, saved)
	if err != nil {
		return pluginInstanceWriteError(e, err)
	}
	return e.JSON(http.StatusOK, public)
}

func PluginSystemInstanceSettings(e *core.RequestEvent) error {
	if e.Auth == nil || e.HasSuperuserAuth() {
		return apis.NewForbiddenError("An authenticated plugin owner is required.", nil)
	}
	dto, err := bindPluginInstanceSettings(e, false)
	if err != nil {
		return err
	}
	id := e.Request.PathValue("id")
	original, err := e.App.FindRecordById("plugin_instances", id)
	if err != nil || original.GetString("user") != e.Auth.Id {
		return apis.NewNotFoundError("Plugin instance not found.", nil)
	}
	c := pluginsystem.CaptureConnection(original)
	if pluginInstanceSettingsBeforeCommit != nil {
		pluginInstanceSettingsBeforeCommit()
	}
	var saved *core.Record
	err = e.App.RunInTransaction(func(tx core.App) error {
		record, err := tx.FindRecordById("plugin_instances", id)
		if err != nil {
			return pluginsystem.ErrConnectionConflict
		}
		if record.GetString("user") != c.UserID || record.GetString("user") != e.Auth.Id || record.GetString("plugin_id") != c.PluginID {
			return pluginsystem.ErrConnectionConflict
		}
		if record.GetString("connection_generation") == "" || record.GetString("oauth_token_revision") == "" {
			return pluginsystem.ErrConnectionConflict
		}
		manifest, err := pluginsystem.InstalledPluginManifest(tx, record.GetString("plugin_id"))
		if err != nil {
			return apis.NewBadRequestError("The installed plugin is unavailable.", nil)
		}
		selected := record.GetString("auth_context")
		if choice, ok := dto["authContext"].(string); ok {
			if len(manifest.Auth.Contexts) == 0 && choice != "" {
				return apis.NewBadRequestError("This plugin has no authentication context.", nil)
			}
			if len(manifest.Auth.Contexts) > 0 {
				if _, exists := manifest.Auth.Contexts[choice]; !exists {
					return apis.NewBadRequestError("Select a valid authentication context.", nil)
				}
			}
			selected = choice
		}
		auth, activeChanged, grantChanged, err := applyClientCredentialPatch(record, manifest, selected, dto["auth"])
		if err != nil {
			return err
		}
		activeChanged = activeChanged || selected != record.GetString("auth_context")
		if activeChanged || clientActiveCredentialIntent(record, manifest, dto["auth"]) {
			expected, _ := dto["expected_connection_generation"].(string)
			if expected == "" {
				return pluginsystem.ConnectionGenerationRequiredError()
			}
			if expected != record.GetString("connection_generation") {
				return pluginsystem.ErrConnectionConflict
			}
		}
		oldConfig := pluginsystem.JSONMapFromRecord(record, "config")
		config := pluginsystem.CloneJSONMap(oldConfig)
		if patch, ok := dto["config"].(map[string]any); ok {
			pluginsystem.MergePluginConfig(config, patch)
		}
		before := installedPluginConfig(tx, c.PluginID)
		pluginsystem.MergePluginConfig(before, oldConfig)
		after := installedPluginConfig(tx, c.PluginID)
		pluginsystem.MergePluginConfig(after, config)
		activation := dto["enabled"] == true && !record.GetBool("enabled")
		if activation || !reflect.DeepEqual(configValidationValue(before), configValidationValue(after)) {
			if _, err := pluginInstancePolicy(pluginsystem.LocalPlugin{Manifest: manifest}, after); err != nil {
				return pluginOperationAPIError(err)
			}
		}
		if _, explicit := dto["enabled"]; explicit {
			auth["oauthActivateOnCallback"] = false
		}
		var previousAuth map[string]any
		previousAuthErr := record.UnmarshalJSONField("auth", &previousAuth)
		if _, explicit := dto["auth"]; explicit || previousAuthErr == nil && previousAuth != nil {
			record.Set("auth", auth)
		}
		if activeChanged {
			pluginsystem.ResetConnection(record)
		}
		if grantChanged {
			record.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
		}
		record.Set("auth_context", selected)
		if _, changed := dto["config"]; changed {
			record.Set("config", config)
		}
		if enabled, explicit := dto["enabled"].(bool); explicit {
			record.Set("enabled", enabled)
		}
		if activation || activeChanged {
			var freshAuth map[string]any
			_ = record.UnmarshalJSONField("auth", &freshAuth)
			status := pluginsystem.LocalConnectionStatus(manifest, selected, freshAuth, record.GetBool("enabled"), time.Now())
			if activation && status != "configured" {
				return apis.NewBadRequestError("The plugin connection is not ready to enable.", nil)
			}
			record.Set("status", status)
		} else if dto["enabled"] == false {
			record.Set("status", "disabled")
		}
		if err := tx.Save(record); err != nil {
			return err
		}
		saved = record.Fresh()
		return nil
	})
	if err != nil {
		return pluginInstanceWriteError(e, err)
	}
	public, err := pluginsystem.PublicConnectionRecord(e.App, saved)
	if err != nil {
		return pluginInstanceWriteError(e, err)
	}
	return e.JSON(http.StatusOK, public)
}

func configValidationValue(config map[string]any) map[string]any {
	result := pluginsystem.CloneJSONMap(config)
	if host, ok := result["host"].(map[string]any); ok {
		delete(host, "categoryRemapDismissedAt")
		if len(host) == 0 {
			delete(result, "host")
		}
	}
	return result
}

// Explicit input is plaintext, even when bytes resemble an existing encrypted
// value. Comparison affects generation; it never decides input encryption.
func applyClientCredentialPatch(record *core.Record, manifest pluginsystem.Manifest, selected string, raw any) (map[string]any, bool, bool, error) {
	auth := pluginsystem.JSONMapFromRecord(record, "auth")
	originalAuth := pluginsystem.CloneJSONMap(auth)
	patch, _ := raw.(map[string]any)
	editable := pluginsystem.EditableAuthCredentialFields(manifest, selected)
	activeFields := map[string]bool{}
	oldContext := manifest.Auth.Contexts[record.GetString("auth_context")]
	for _, field := range append(append([]string{}, oldContext.Fields...), oldContext.SecretFields...) {
		activeFields[field] = true
	}
	if oldContext.SecretField != "" {
		activeFields[oldContext.SecretField] = true
	}
	if oldContext.Type == pluginsystem.AuthTypeOAuth2 {
		activeFields["clientSecret"] = true
	}
	secrets := pluginsystem.ManifestAuthSecretFields(manifest)
	activeChanged, grantChanged := false, false
	var explicitAccessToken any
	// API-key/Bearer contexts may declare accessToken as an input. Reusing
	// that storage key consumes the old OAuth grant before the explicit input
	// is applied; selecting OAuth never makes its grant directly editable.
	if input, exists := patch["accessToken"]; exists && editable["accessToken"] {
		if text, ok := input.(string); input == nil || ok && text != "" {
			pluginsystem.ClearOAuthGrant(auth)
			grantChanged = true
		}
	}
	for key, value := range patch {
		if !editable[key] {
			return nil, false, false, apis.NewBadRequestError("Unsupported credential field.", nil)
		}
		text, isString := value.(string)
		if value != nil && !isString {
			return nil, false, false, apis.NewBadRequestError("Credentials must be strings or null.", nil)
		}
		if value == nil && !secrets[key] {
			return nil, false, false, apis.NewBadRequestError("Only editable secret credentials can be removed with null.", nil)
		}
		if secrets[key] && text == "" && value != nil {
			continue
		}
		prior, existed := originalAuth[key]
		changed := existed && value == nil || !existed && value != nil
		if value != nil && existed {
			old, ok := prior.(string)
			if secrets[key] && ok && old != "" && util.LooksLikeEncrypted(old) {
				decoded, err := security.Decrypt(old, os.Getenv("POCKETBASE_ENCRYPTION_KEY"))
				if err == nil {
					old = string(decoded)
				} else if !pluginsystem.IsLegacyPlaintextAuthSecret(manifest.ID, key, old) {
					ok = false
				}
			}
			changed = !ok || old != text
		}
		if changed && activeFields[key] {
			activeChanged = true
		}
		if value == nil {
			delete(auth, key)
		} else if secrets[key] && text != "" {
			encrypted, err := security.Encrypt([]byte(text), os.Getenv("POCKETBASE_ENCRYPTION_KEY"))
			if err != nil {
				return nil, false, false, err
			}
			auth[key] = encrypted
		} else {
			auth[key] = text
		}
		if key == "accessToken" && value != nil {
			explicitAccessToken = auth[key]
		}
		if changed {
			grantContext, _ := originalAuth["oauthContext"].(string)
			grant := manifest.Auth.Contexts[grantContext]
			// A legacy grant without provenance is compatible with a sole OAuth
			// context. Its client inputs still own that grant; changing them
			// must consume the tokens exactly as for an attributed grant.
			if provenance, exists := originalAuth["oauthContext"]; (!exists || provenance == "") && len(manifest.Auth.Contexts) == 1 {
				_, sole, err := pluginsystem.ResolveAuthContext(manifest, "")
				if err == nil && sole.Type == pluginsystem.AuthTypeOAuth2 {
					grant = sole
				}
			}
			if grant.Type == pluginsystem.AuthTypeOAuth2 {
				for _, field := range append(append([]string{}, grant.Fields...), "clientSecret") {
					if field == key && key != "accessToken" && key != "refreshToken" {
						pluginsystem.ClearOAuthGrant(auth)
						grantChanged = true
						break
					}
				}
			}
			if key == "accessToken" || key == "refreshToken" {
				grantChanged = true
			}
		}
	}
	if explicitAccessToken != nil {
		auth["accessToken"] = explicitAccessToken
	}
	return auth, activeChanged, grantChanged, nil
}

func pluginInstanceWriteError(e *core.RequestEvent, err error) error {
	if errors.Is(err, pluginsystem.ErrConnectionObsolete) || errors.Is(err, pluginsystem.ErrInstanceExists) {
		return pluginsystem.ConnectionAPIError(err)
	}
	var apiError *router.ApiError
	if errors.As(err, &apiError) {
		return apiError
	}
	context := pluginOperationLogContext{Action: pluginActionInstanceCreate}
	if id := e.Request.PathValue("id"); id != "" {
		context.Action = pluginActionInstanceSettings
		if record, lookupErr := e.App.FindRecordById("plugin_instances", id); lookupErr == nil && e.Auth != nil && record.GetString("user") == e.Auth.Id {
			context.PluginID, context.InstanceID = record.GetString("plugin_id"), record.Id
		}
	}
	return pluginHostAPIErrorForEvent(e, context, "The plugin instance could not be saved.", pluginsystem.WithDiagnosticCause(err, "database_save"))
}

func clientActiveCredentialIntent(record *core.Record, manifest pluginsystem.Manifest, raw any) bool {
	patch, _ := raw.(map[string]any)
	context := manifest.Auth.Contexts[record.GetString("auth_context")]
	fields := append(append([]string{}, context.Fields...), context.SecretFields...)
	if context.SecretField != "" {
		fields = append(fields, context.SecretField)
	}
	if context.Type == pluginsystem.AuthTypeOAuth2 {
		fields = append(fields, "clientSecret")
	}
	secrets, auth := pluginsystem.ManifestAuthSecretFields(manifest), pluginsystem.JSONMapFromRecord(record, "auth")
	for _, key := range fields {
		value, exists := patch[key]
		if !exists {
			continue
		}
		if value == nil {
			if _, stored := auth[key]; stored {
				return true
			}
			continue
		}
		if text, ok := value.(string); ok && (!secrets[key] || text != "") {
			return true
		}
	}
	return false
}
