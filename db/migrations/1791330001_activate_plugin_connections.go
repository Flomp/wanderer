package migrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"pocketbase/pluginsystem"
)

func init() { m.Register(activatePluginConnections1791330001, nil) }

func activatePluginConnections1791330001(app core.App) error {
	return activatePluginConnectionsWithManifestDir(app, pluginsystem.PluginDir())
}

// Older upgrades can populate plugin instances before the first discovery has
// cached their manifests. Read JSON only; migration must not load/execute WASM.
func connectionMigrationManifests(dir string) (map[string]pluginsystem.Manifest, error) {
	manifests := map[string]pluginsystem.Manifest{}
	ambiguous := map[string]bool{}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return manifests, nil
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name(), "plugin.json"))
		if err != nil {
			continue
		}
		var manifest pluginsystem.Manifest
		if json.Unmarshal(data, &manifest) != nil || pluginsystem.ValidateManifest(manifest) != nil {
			continue
		}
		if _, exists := manifests[manifest.ID]; exists || ambiguous[manifest.ID] {
			delete(manifests, manifest.ID)
			ambiguous[manifest.ID] = true
			continue
		}
		manifests[manifest.ID] = manifest
	}
	return manifests, nil
}

func activatePluginConnectionsWithManifestDir(app core.App, dir string) error {
	now := time.Now()
	return app.RunInTransaction(func(tx core.App) error {
		var fallback map[string]pluginsystem.Manifest
		fallbackLoaded := false
		collection, err := tx.FindCollectionByNameOrId("plugin_instances")
		if err != nil {
			return err
		}
		records, err := tx.FindAllRecords(collection)
		if err != nil {
			return err
		}
		for _, record := range records {
			if record.GetString("connection_generation") != "" {
				continue
			}
			before := map[string]string{}
			for _, field := range []string{"auth_context", "connection_generation", "oauth_token_revision", "status", "auth"} {
				before[field] = record.GetString(field)
			}
			manifest, manifestErr := pluginsystem.InstalledPluginManifest(tx, record.GetString("plugin_id"))
			if manifestErr != nil && !fallbackLoaded {
				fallback, err = connectionMigrationManifests(dir)
				if err != nil {
					// An unavailable bundle directory must not prevent the host
					// from starting. Leave selection unresolved for owner repair.
					tx.Logger().Warn("plugin connection migration could not read bundle manifests", "cause", "bundle_directory_unreadable")
					fallback = nil
				}
				fallbackLoaded = true
			}
			if local, exists := fallback[record.GetString("plugin_id")]; manifestErr != nil && exists {
				manifest, manifestErr = local, nil
			}
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
