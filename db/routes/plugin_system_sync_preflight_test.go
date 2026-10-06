package routes

import (
	"context"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

func TestCronInvalidEffectiveConfigDoesNotRefreshExpiredGrant(t *testing.T) {
	for _, test := range []struct {
		name      string
		installed bool
		config    map[string]any
	}{
		{name: "missing configured connector"},
		{name: "invalid stored connector base", config: map[string]any{"host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": "file:///private-config"}}}}},
		{name: "changed installation default", installed: true, config: map[string]any{"host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": "https://[invalid-config"}}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			for i := range f.plugin.Manifest.Permissions.Network.Connectors {
				connector := &f.plugin.Manifest.Permissions.Network.Connectors[i]
				if connector.Name == "api" {
					connector.Type = pluginsystem.ConnectorTypeConfigured
					connector.ConfigKey = "server"
					connector.FixedBaseURL = ""
				}
			}
			installed, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", f.plugin.Manifest.ID)
			if err != nil {
				t.Fatal(err)
			}
			installed.Set("manifest", f.plugin.Manifest)
			if test.installed {
				installed.Set("config", test.config)
			}
			if err := f.app.Save(installed); err != nil {
				t.Fatal(err)
			}
			collection, err := f.app.FindCollectionByNameOrId("plugin_instances")
			if err != nil {
				t.Fatal(err)
			}
			collection.Fields.Add(&core.DateField{Name: "retry_not_before"})
			if err := f.app.Save(collection); err != nil {
				t.Fatal(err)
			}
			instance, auth := f.storedAuth(t)
			if !pluginsystem.OAuthNeedsRefresh(auth) {
				t.Fatal("fixture must require a real token refresh")
			}
			if !test.installed && test.config != nil {
				instance.Set("config", test.config)
				if err := f.app.Save(instance); err != nil {
					t.Fatal(err)
				}
			}
			actors := core.NewBaseCollection("activitypub_actors")
			actors.Fields.Add(&core.TextField{Name: "user"})
			if err := f.app.Save(actors); err != nil {
				t.Fatal(err)
			}
			actor := core.NewRecord(actors)
			actor.Set("user", instance.GetString("user"))
			if err := f.app.Save(actor); err != nil {
				t.Fatal(err)
			}
			fresh, err := f.app.FindRecordById("plugin_instances", instance.Id)
			if err != nil {
				t.Fatal(err)
			}
			beforeAuth := fresh.GetString("auth")
			_, err = syncPluginInstance(context.Background(), f.app, nil, f.plugin, fresh)
			if err == nil || pluginsystem.PublicPluginError(err).Code != "invalid_configuration" {
				t.Fatalf("invalid config was not rejected locally: %v", err)
			}
			if f.tokenRequests.Load() != 0 {
				t.Fatal("invalid config reached OAuth provider and rotated the grant")
			}
			fresh, err = f.app.FindRecordById("plugin_instances", instance.Id)
			if err != nil {
				t.Fatal(err)
			}
			if fresh.GetString("auth") != beforeAuth {
				t.Fatal("invalid config changed persisted credentials/grant")
			}
			if fresh.GetString("status") != "error" || pluginsystem.JSONMapFromRecord(fresh, "last_error")["code"] != "invalid_configuration" {
				t.Fatalf("missing local configuration diagnosis: %#v", pluginsystem.JSONMapFromRecord(fresh, "last_error"))
			}
		})
	}
}
