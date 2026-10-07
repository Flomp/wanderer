package pluginsystem

import (
	"errors"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// InstalledPluginManifest reads the database cache only. Read projection and
// callers can read it without discovering bundles or executing WASM.
func InstalledPluginManifest(app core.App, id string) (Manifest, error) {
	record, err := app.FindFirstRecordByFilter("installed_plugins", "plugin_id={:id}", dbx.Params{"id": id})
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := record.UnmarshalJSONField("manifest", &manifest); err != nil {
		return Manifest{}, err
	}
	if manifest.ID != id {
		return Manifest{}, errors.New("installed plugin manifest mismatch")
	}
	if err := ValidateManifest(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

var publicConnectionFields = map[string]bool{
	"id": true, "user": true, "plugin_id": true, "enabled": true, "auth": true,
	"config": true, "state": true, "status": true, "last_error": true,
	"last_sync_at": true, "retry_not_before": true, "created": true, "updated": true,
	"auth_context": true, "connection_generation": true,
}

var privateConnectionKeys = map[string]bool{
	"oauth_token_revision": true, "oauthTokenRevision": true,
	"oauthState": true, "oauthCodeVerifier": true, "oauthRedirectURI": true,
	"oauthFlowContext": true, "oauthActivateOnCallback": true,
}

// ProjectConnectionRecord is the common read boundary, applied after standard
// PocketBase enrichment. It never saves the record or decrypts a value.
func ProjectConnectionRecord(app core.App, record *core.Record, owner bool) {
	for _, field := range record.Collection().Fields {
		if !publicConnectionFields[field.GetName()] {
			record.Hide(field.GetName())
		}
	}
	for key := range record.CustomData() {
		if key != "busy" {
			record.Hide(key)
		}
	}
	manifest, err := InstalledPluginManifest(app, record.GetString("plugin_id"))
	secrets := ManifestAuthSecretFields(manifest)
	auth := map[string]any{}
	if err == nil {
		allowed := map[string]bool{"accessToken": true, "refreshToken": true, "expiresAt": true, "tokenType": true, "scope": true, "oauthContext": true, "providerUserId": true}
		for _, ctx := range manifest.Auth.Contexts {
			if ctx.Type == AuthTypeOAuth2 {
				allowed["clientSecret"] = true
			}
			for _, key := range append(append([]string{}, ctx.Fields...), ctx.SecretFields...) {
				allowed[key] = true
			}
			if ctx.SecretField != "" {
				allowed[ctx.SecretField] = true
			}
		}
		for key, value := range JSONMapFromRecord(record, "auth") {
			if !allowed[key] || privateConnectionKeys[key] {
				continue
			}
			if secrets[key] {
				if text, ok := value.(string); ok && text != "" {
					auth[key] = ""
				}
				continue
			}
			auth[key] = redactConnectionJSON(value, secrets)
		}
	}
	record.Set("auth", auth)
	for _, field := range []string{"config", "state", "last_error"} {
		record.Set(field, redactConnectionJSON(JSONMapFromRecord(record, field), secrets))
	}
	record.WithCustomData(true)
	if owner {
		record.Set("busy", InstanceBusy(record.Id))
		record.Unhide("busy")
	} else {
		record.Hide("busy")
	}
}

func redactConnectionJSON(value any, secrets map[string]bool) any {
	switch typed := value.(type) {
	case map[string]any:
		result := map[string]any{}
		for key, child := range typed {
			if privateConnectionKeys[key] || secrets[key] {
				continue
			}
			result[key] = redactConnectionJSON(child, secrets)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, child := range typed {
			result[i] = redactConnectionJSON(child, secrets)
		}
		return result
	default:
		return CloneJSONValue(value)
	}
}

// Fresh is shallow: projected JSON values are separately rebuilt so host
// response redaction never modifies a save writer's working record.
func PublicConnectionRecord(app core.App, record *core.Record) (*core.Record, error) {
	copy := record.Fresh()
	ProjectConnectionRecord(app, copy, true)
	return copy, nil
}
