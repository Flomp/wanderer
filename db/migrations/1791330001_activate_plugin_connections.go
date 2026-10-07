package migrations

import (
	"encoding/json"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"pocketbase/pluginsystem"
)

func init() { m.Register(activatePluginConnections1791330001, nil) }

func activatePluginConnections1791330001(app core.App) error {
	now := time.Now()
	return app.RunInTransaction(func(tx core.App) error {
		collection, err := tx.FindCollectionByNameOrId("plugin_instances")
		if err != nil {
			return err
		}
		records, err := tx.FindAllRecords(collection)
		if err != nil {
			return err
		}
		for _, record := range records {
			before := map[string]string{}
			for _, field := range []string{"auth_context", "connection_generation", "oauth_token_revision", "status", "auth"} {
				before[field] = record.GetString(field)
			}
			manifest, manifestErr := pluginsystem.InstalledPluginManifest(tx, record.GetString("plugin_id"))
			var auth map[string]any
			authErr := json.Unmarshal([]byte(record.GetString("auth")), &auth)
			selected := record.GetString("auth_context")
			if selected == "" && record.GetString("connection_generation") == "" && authErr == nil {
				legacy, exists := auth["oauthContext"]
				legacyText, stringValue := legacy.(string)
				if stringValue && legacyText != "" {
					selected = legacyText
				} else if (!exists || stringValue) && manifestErr == nil && len(manifest.Auth.Contexts) == 1 {
					for name := range manifest.Auth.Contexts {
						selected = name
					}
				}
			}
			record.Set("auth_context", selected)
			if record.GetString("connection_generation") == "" {
				record.Set("connection_generation", pluginsystem.NewConnectionRevision())
			}
			if record.GetString("oauth_token_revision") == "" {
				record.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
			}
			if record.GetString("status") == "syncing" {
				status := "error"
				if !record.GetBool("enabled") {
					status = "disabled"
				} else if manifestErr == nil {
					status = pluginsystem.LocalConnectionStatus(manifest, selected, auth, true, now)
				}
				record.Set("status", status)
			}
			if authErr == nil && auth != nil {
				pluginsystem.ClearPendingOAuthFlow(auth)
				record.Set("auth", auth)
			}
			changes := dbx.Params{}
			for field, prior := range before {
				if record.GetString(field) != prior {
					changes[field] = record.GetString(field)
				}
			}
			// A data migration neither runs record hooks nor updates historical
			// timestamps. Preserve even diagnostically malformed legacy values.
			if len(changes) > 0 {
				if _, err := tx.DB().Update(collection.Name, changes, dbx.HashExp{"id": record.Id}).Execute(); err != nil {
					return err
				}
			}
		}
		collection.CreateRule = nil
		collection.UpdateRule = nil
		return tx.Save(collection)
	})
}
