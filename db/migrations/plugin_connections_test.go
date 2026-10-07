package migrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/pocketbase/pocketbase/tools/types"
	"pocketbase/pluginsystem"
)

func TestPluginConnectionColdUpgradeUsesBundleJSONAndPreservesLiveReapply(t *testing.T) {
	t.Setenv("ORIGIN", "https://cold-upgrade.example.test")
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", "")
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	seeded := false
	for _, migration := range slices.Clone(core.AppMigrations.Items()) {
		if migration.File >= "1791330000_plugin_connection_fields.go" {
			continue
		}
		switch migration.File {
		case "1742167033_init_meilisearch.go", "1744651602_add_polyline.go", "1749831369_update_sortable_attributes.go":
			continue
		}
		if migration.File == "1780000002_plugin_instances.go" {
			users, err := app.FindCollectionByNameOrId("users")
			if err != nil {
				t.Fatal(err)
			}
			user := core.NewRecord(users)
			user.Load(map[string]any{"username": "cold-upgrade", "email": "cold-upgrade@example.test", "password": "cold-upgrade-password"})
			if err := app.Save(user); err != nil {
				t.Fatal(err)
			}
			integrations, err := app.FindCollectionByNameOrId("integrations")
			if err != nil {
				t.Fatal(err)
			}
			legacy := core.NewRecord(integrations)
			legacy.Set("user", user.Id)
			legacy.Set("strava", map[string]any{"active": true, "clientId": "legacy-client", "clientSecret": "legacy-secret", "accessToken": strings.Repeat("a", 40), "refreshToken": strings.Repeat("b", 40), "expiresAt": "2099-01-01T00:00:00Z", "activities": true, "routes": true})
			for _, id := range []string{"komoot", "hammerhead"} {
				legacy.Set(id, map[string]any{"active": true, "email": "legacy@example.test", "password": "legacy-password", "completed": true, "planned": true})
			}
			if err := app.Save(legacy); err != nil {
				t.Fatal(err)
			}
			seeded = true
		}
		if err := app.RunInTransaction(migration.Up); err != nil {
			t.Fatalf("released migration %s: %v", migration.File, err)
		}
	}
	if !seeded {
		t.Fatal("fixture never exercised legacy integration migration")
	}
	if count, err := app.CountRecords("installed_plugins"); err != nil || count != 0 {
		t.Fatalf("cold cache not empty: %d/%v", count, err)
	}
	rows, err := app.FindAllRecords("plugin_instances")
	if err != nil || len(rows) != 3 {
		t.Fatalf("legacy instances=%v/%v", rows, err)
	}
	before := map[string]map[string]any{}
	for _, row := range rows {
		if row.GetString("plugin_id") == "strava" {
			row.Set("status", "syncing")
		}
		row.Set("state", map[string]any{"cursor": "preserved-cursor"})
		row.Set("last_error", map[string]any{"code": "rate_limited", "historical": true})
		row.Set("retry_not_before", "2099-02-03 04:05:06.000Z")
		row.Set("last_sync_at", "2020-02-03 04:05:06.000Z")
		if err := app.Save(row); err != nil {
			t.Fatal(err)
		}
		row, err = app.FindRecordById("plugin_instances", row.Id)
		if err != nil {
			t.Fatal(err)
		}
		before[row.Id] = row.FieldsData()
	}
	dir := t.TempDir()
	selected := map[string]string{}
	for _, id := range []string{"strava", "komoot", "hammerhead"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "plugins", id, "plugin.json"))
		if err != nil {
			t.Fatal(err)
		}
		var manifest pluginsystem.Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			t.Fatal(err)
		}
		if len(manifest.Auth.Contexts) != 1 {
			t.Fatalf("released %s fixture requires exactly one context", id)
		}
		for name := range manifest.Auth.Contexts {
			selected[id] = name
		}
		bundle := filepath.Join(dir, "directory-"+id)
		if err := os.MkdirAll(bundle, 0700); err != nil {
			t.Fatal(err)
		}
		// Deliberately omit the WASM entrypoint: this fallback reads only JSON.
		if err := os.WriteFile(filepath.Join(bundle, "plugin.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := addPluginConnectionFields1791330000(app); err != nil {
		t.Fatal(err)
	}
	if err := activatePluginConnectionsWithManifestDir(app, dir); err != nil {
		t.Fatal(err)
	}
	for _, old := range rows {
		row, err := app.FindRecordById("plugin_instances", old.Id)
		if err != nil {
			t.Fatal(err)
		}
		if row.GetString("auth_context") != selected[row.GetString("plugin_id")] || row.GetString("connection_generation") == "" || row.GetString("oauth_token_revision") == "" {
			t.Fatalf("cold upgrade %s selection=%q, expected %q; G/T present=%t/%t", row.GetString("plugin_id"), row.GetString("auth_context"), selected[row.GetString("plugin_id")], row.GetString("connection_generation") != "", row.GetString("oauth_token_revision") != "")
		}
		for field, value := range before[row.Id] {
			if field == "status" && value == "syncing" {
				value = "configured"
			}
			if !reflect.DeepEqual(value, row.Get(field)) {
				t.Errorf("cold upgrade changed %s.%s", row.GetString("plugin_id"), field)
			}
		}
		// A later migration run must not clear a newly started encrypted flow or
		// reconstruct status/state/diagnoses on an already initialized instance.
		auth := pluginsystem.JSONMapFromRecord(row, "auth")
		auth["oauthState"] = "current-encrypted-state"
		auth["oauthCodeVerifier"] = "current-encrypted-verifier"
		auth["oauthRedirectURI"] = "https://current.example/callback"
		auth["oauthFlowContext"] = row.GetString("auth_context")
		auth["oauthActivateOnCallback"] = true
		row.Set("auth", auth)
		row.Set("status", "syncing")
		if err := app.Save(row); err != nil {
			t.Fatal(err)
		}
		row, err = app.FindRecordById("plugin_instances", row.Id)
		if err != nil {
			t.Fatal(err)
		}
		before[row.Id] = row.FieldsData()
	}
	if err := activatePluginConnectionsWithManifestDir(app, dir); err != nil {
		t.Fatal(err)
	}
	for id, fields := range before {
		row, err := app.FindRecordById("plugin_instances", id)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(fields, row.FieldsData()) {
			t.Fatalf("reapplied migration changed live instance %s", id)
		}
	}
	// Reapply does not even read bundles when every record already has G.
	noDirectory := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(noDirectory, []byte("not a plugin directory"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := activatePluginConnectionsWithManifestDir(app, noDirectory); err != nil {
		t.Fatal(err)
	}
	if count, err := app.CountRecords("installed_plugins"); err != nil || count != 0 {
		t.Fatalf("migration ran discovery: %d/%v", count, err)
	}
}

func TestConnectionMigrationManifestFallbackRejectsAmbiguousOrInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	valid := migrationConnectionManifest("duplicate", map[string]pluginsystem.AuthContext{"api": {Type: pluginsystem.AuthTypeAPIKey, SecretField: "apiKey"}})
	for name, value := range map[string]any{"first": valid, "second": valid, "invalid": map[string]any{"id": "invalid"}, "single": migrationConnectionManifest("single", nil)} {
		bundle := filepath.Join(dir, name)
		if err := os.MkdirAll(bundle, 0700); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(bundle, "plugin.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	manifests, err := connectionMigrationManifests(dir)
	if err != nil || len(manifests) != 1 || manifests["single"].ID != "single" {
		t.Fatalf("unsafe fallback: %#v/%v", manifests, err)
	}
}

func migrationConnectionManifest(id string, contexts map[string]pluginsystem.AuthContext) pluginsystem.Manifest {
	return pluginsystem.Manifest{ManifestVersion: "1.0", ID: id, Type: pluginsystem.PluginTypeTrails, Name: id, Version: "1.0", Runtime: pluginsystem.RuntimeManifest{Type: pluginsystem.RuntimeWASM, Entrypoint: "plugin.wasm"}, Capabilities: []pluginsystem.CapabilityManifest{{Name: "list", Version: "1.0", Export: "list"}}, Auth: pluginsystem.AuthManifest{Contexts: contexts}}
}

func TestActivatePluginConnectionsMigratesSelectionAndSyncingLocally(t *testing.T) {
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	instances := core.NewBaseCollection("plugin_instances")
	instances.CreateRule = types.Pointer("")
	instances.UpdateRule = types.Pointer("")
	for _, field := range []string{"plugin_id", "status", "user", "last_sync_at", "retry_not_before", "updated"} {
		instances.Fields.Add(&core.TextField{Name: field})
	}
	for _, field := range []string{"auth", "config", "state", "last_error"} {
		instances.Fields.Add(&core.JSONField{Name: field})
	}
	instances.Fields.Add(&core.BoolField{Name: "enabled"})
	if err := app.Save(instances); err != nil {
		t.Fatal(err)
	}
	if err := addPluginConnectionFields1791330000(app); err != nil {
		t.Fatal(err)
	}
	instances, _ = app.FindCollectionByNameOrId("plugin_instances")
	installed := core.NewBaseCollection("installed_plugins")
	installed.Fields.Add(&core.TextField{Name: "plugin_id"}, &core.JSONField{Name: "manifest"})
	if err := app.Save(installed); err != nil {
		t.Fatal(err)
	}
	api := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeAPIKey, SecretField: "apiKey"}
	session := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeSession, Fields: []string{"email", "password"}, SecretFields: []string{"password"}, Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModePlugin, Function: "refresh_auth_v1"}}
	oauth := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeOAuth2, Fields: []string{"clientId", "clientSecret"}, AuthorizationURL: "https://example.test/auth", TokenURL: "https://example.test/token", Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModeHost}}
	for id, contexts := range map[string]map[string]pluginsystem.AuthContext{"single": {"api": api}, "multi": {"api": api, "other": api}, "session": {"web": session}, "oauth": {"oauth": oauth}, "none": nil} {
		r := core.NewRecord(installed)
		r.Set("plugin_id", id)
		r.Set("manifest", migrationConnectionManifest(id, contexts))
		if err := app.Save(r); err != nil {
			t.Fatal(err)
		}
	}
	type fixture struct {
		name, plugin, selected, status string
		auth                           any
		enabled                        bool
	}
	legacyCiphertext := func(value string) string {
		encrypted, err := security.Encrypt([]byte(value), "0123456789abcdef0123456789abcdef")
		if err != nil {
			t.Fatal(err)
		}
		return encrypted
	}
	legacyFlowAuth := map[string]any{
		"clientId": "legacy-client", "clientSecret": legacyCiphertext("client-secret"),
		"accessToken": legacyCiphertext("grant-access"), "refreshToken": legacyCiphertext("grant-refresh"),
		"oauthContext": "oauth", "expiresAt": "2099-01-01T00:00:00Z", "providerUserId": "legacy-account",
		"oauthState": legacyCiphertext("pending-state"), "oauthCodeVerifier": legacyCiphertext("pending-verifier"),
		"oauthRedirectURI": "https://old.example/callback", "oauthFlowContext": "oauth", "oauthActivateOnCallback": true,
	}
	cases := []fixture{
		{"single absent", "single", "api", "configured", map[string]any{"apiKey": "opaque"}, true},
		{"single empty", "single", "api", "needs_auth", map[string]any{"oauthContext": ""}, true},
		{"single null", "single", "", "error", map[string]any{"oauthContext": nil, "apiKey": "opaque"}, true},
		{"single wrong legacy", "single", "", "error", map[string]any{"oauthContext": 42}, true},
		{"known", "multi", "api", "configured", map[string]any{"oauthContext": "api", "apiKey": "opaque"}, true},
		{"unknown", "single", "unknown", "error", map[string]any{"oauthContext": "unknown", "apiKey": "opaque"}, true},
		{"multi unresolved", "multi", "", "error", map[string]any{"apiKey": "opaque"}, true},
		{"no auth", "none", "", "configured", nil, true},
		{"missing manifest", "missing", "", "error", map[string]any{}, true},
		{"disabled echo", "oauth", "oauth", "disabled", map[string]any{"accessToken": "opaque"}, false},
		{"session complete", "session", "web", "configured", map[string]any{"email": "e", "password": "ciphertext"}, true},
		{"session missing", "session", "web", "needs_auth", map[string]any{"email": "e"}, true},
		{"credential wrong type", "single", "api", "error", map[string]any{"apiKey": 123}, true},
		{"credential null", "single", "api", "error", map[string]any{"apiKey": nil}, true},
		{"auth not object", "single", "", "error", []any{"opaque"}, true},
		{"oauth absent", "oauth", "oauth", "needs_auth", map[string]any{}, true},
		{"oauth unknown expiry", "oauth", "oauth", "configured", map[string]any{"accessToken": "opaque", "expiresAt": "invalid"}, true},
		{"oauth future", "oauth", "oauth", "configured", map[string]any{"accessToken": "opaque", "expiresAt": time.Now().Add(time.Hour).Format(time.RFC3339)}, true},
		{"oauth expired", "oauth", "oauth", "needs_reauth", map[string]any{"accessToken": "opaque", "expiresAt": "2000-01-01T00:00:00Z"}, true},
		{"oauth refresh-only", "oauth", "oauth", "configured", map[string]any{"refreshToken": "opaque", "clientId": "client", "clientSecret": "cipher"}, true},
		{"oauth missing refresh prerequisite", "oauth", "oauth", "needs_reauth", map[string]any{"refreshToken": "opaque"}, true},
		{"successful refresh then crash", "oauth", "oauth", "configured", map[string]any{"accessToken": "new-grant", "refreshToken": "new-refresh", "expiresAt": "2099-01-01T00:00:00Z"}, true},
		{"non-syncing diagnosis", "oauth", "oauth", "needs_reauth", map[string]any{"accessToken": "present-grant", "expiresAt": "2099-01-01T00:00:00Z"}, true},
		{"non-syncing legacy flow", "oauth", "oauth", "configured", legacyFlowAuth, true},
	}
	records := map[string]*core.Record{}
	for _, c := range cases {
		r := core.NewRecord(instances)
		r.Set("plugin_id", c.plugin)
		r.Set("status", "syncing")
		if c.name == "non-syncing diagnosis" || c.name == "non-syncing legacy flow" {
			r.Set("status", c.status)
		}
		r.Set("enabled", c.enabled)
		r.Set("auth", c.auth)
		r.Set("config", map[string]any{"plugin": map[string]any{"after": "old"}})
		r.Set("state", map[string]any{"cursor": 42})
		r.Set("last_error", map[string]any{"code": "auth_failed", "arbitrary": "historical"})
		r.Set("retry_not_before", "2000-02-03")
		r.Set("last_sync_at", "2000-01-01")
		r.Set("updated", "2000-01-01")
		if err := app.Save(r); err != nil {
			t.Fatal(err)
		}
		records[c.name] = r
	}
	if err := activatePluginConnections1791330001(app); err != nil {
		t.Fatal(err)
	}
	first := map[string]map[string]any{}
	for _, c := range cases {
		r, err := app.FindRecordById("plugin_instances", records[c.name].Id)
		if err != nil {
			t.Fatal(err)
		}
		if r.GetString("auth_context") != c.selected || r.GetString("status") != c.status {
			t.Errorf("%s selection/status=%q/%q expected %q/%q", c.name, r.GetString("auth_context"), r.GetString("status"), c.selected, c.status)
		}
		if r.GetString("connection_generation") == "" || r.GetString("oauth_token_revision") == "" {
			t.Errorf("%s missing revisions", c.name)
		}
		if c.name == "non-syncing legacy flow" {
			expectedAuth := pluginsystem.JSONMapFromRecord(records[c.name], "auth")
			for _, field := range []string{"oauthState", "oauthCodeVerifier", "oauthRedirectURI", "oauthFlowContext", "oauthActivateOnCallback"} {
				delete(expectedAuth, field)
			}
			if actual := pluginsystem.JSONMapFromRecord(r, "auth"); !reflect.DeepEqual(actual, expectedAuth) {
				t.Errorf("legacy pending flow removal changed grant/client material: %#v", actual)
			}
		} else if !reflect.DeepEqual(r.Get("auth"), records[c.name].Get("auth")) {
			t.Errorf("%s changed preserved auth", c.name)
		}
		for _, field := range []string{"enabled", "config", "state", "last_error", "retry_not_before", "last_sync_at", "updated"} {
			if !reflect.DeepEqual(r.Get(field), records[c.name].Get(field)) {
				t.Errorf("%s changed preserved %s", c.name, field)
			}
		}
		first[c.name] = r.FieldsData()
	}
	collection, _ := app.FindCollectionByNameOrId("plugin_instances")
	if collection.CreateRule != nil || collection.UpdateRule != nil {
		t.Fatal("direct write rules unlocked")
	}
	if err := activatePluginConnections1791330001(app); err != nil {
		t.Fatal(err)
	}
	for name, before := range first {
		r, _ := app.FindRecordById("plugin_instances", records[name].Id)
		if !reflect.DeepEqual(before, r.FieldsData()) {
			t.Errorf("%s second migration changed record", name)
		}
	}
}

// Exercise the actual released application migration definitions, rather than
// rebuilding a subset of the plugin schema by hand. Only migrations dedicated
// to external search index configuration are omitted from this local fixture.
func TestPluginConnectionUpgradeFromReleasedSchemaPreservesFRecords(t *testing.T) {
	t.Setenv("ORIGIN", "https://upgrade.example.test")
	legacyKey := "0123456789abcdef0123456789abcdef"
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	applied := map[string]bool{}
	for _, migration := range slices.Clone(core.AppMigrations.Items()) {
		if migration.File >= "1791330000_plugin_connection_fields.go" {
			continue
		}
		switch migration.File {
		case "1742167033_init_meilisearch.go", "1744651602_add_polyline.go", "1749831369_update_sortable_attributes.go":
			continue
		}
		if err := app.RunInTransaction(migration.Up); err != nil {
			t.Fatalf("released migration %s: %v", migration.File, err)
		}
		applied[migration.File] = true
	}
	for _, name := range []string{"1780000002_plugin_instances.go", "1780000004_plugin_system.go", "1791110001_immutable_record_ownership.go"} {
		if !applied[name] {
			t.Fatalf("fixture omitted released migration %s", name)
		}
	}
	instances, err := app.FindCollectionByNameOrId("plugin_instances")
	if err != nil {
		t.Fatal(err)
	}
	installed, err := app.FindCollectionByNameOrId("installed_plugins")
	if err != nil {
		t.Fatal(err)
	}
	// Prove this is the old schema, including its real date/select/relation
	// fields, original fixed collection IDs, unique indexes and owner rules.
	if instances.Id != "pbc_430001000" || installed.Id != "pbc_430002000" {
		t.Fatal("fixture did not use released plugin collections")
	}
	if instances.CreateRule == nil || instances.UpdateRule == nil {
		t.Fatal("released API writes were already locked")
	}
	for _, field := range []string{"auth_context", "connection_generation", "oauth_token_revision"} {
		if instances.Fields.GetByName(field) != nil {
			t.Fatalf("pre-upgrade schema already has %s", field)
		}
	}
	if _, ok := instances.Fields.GetByName("status").(*core.SelectField); !ok {
		t.Fatal("released status select field missing")
	}
	if _, ok := instances.Fields.GetByName("updated").(*core.AutodateField); !ok {
		t.Fatal("released updated autodate field missing")
	}
	if _, ok := instances.Fields.GetByName("user").(*core.RelationField); !ok {
		t.Fatal("released owner relation missing")
	}
	oldFields := map[string][]byte{}
	for _, field := range instances.Fields {
		oldFields[field.GetName()], err = json.Marshal(field)
		if err != nil {
			t.Fatal(err)
		}
	}
	oldIndexes := slices.Clone(instances.Indexes)
	oldRules := map[string]*string{"list": instances.ListRule, "view": instances.ViewRule, "delete": instances.DeleteRule}
	installedBefore, err := json.Marshal(installed)
	if err != nil {
		t.Fatal(err)
	}
	manifestData, err := os.ReadFile(filepath.Join("..", "..", "plugins", "strava", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest pluginsystem.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	if err := pluginsystem.ValidateManifest(manifest); err != nil {
		t.Fatal(err)
	}
	context := "oauth_access_token"
	if _, exists := manifest.Auth.Contexts[context]; !exists {
		t.Fatal("shipped release OAuth context missing")
	}
	plugin := core.NewRecord(installed)
	plugin.Load(map[string]any{"plugin_id": manifest.ID, "name": manifest.Name, "type": manifest.Type, "version": manifest.Version, "runtime": manifest.Runtime.Type, "manifest": manifest, "config": map[string]any{"plugin": map[string]any{"after": "2020-01-01"}}, "status": "available"})
	if err := app.Save(plugin); err != nil {
		t.Fatal(err)
	}
	plugin, err = app.FindRecordById(installed.Id, plugin.Id)
	if err != nil {
		t.Fatal(err)
	}
	pluginBefore := plugin.FieldsData()
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatal(err)
	}
	encrypt := func(value string) string {
		encrypted, err := security.Encrypt([]byte(value), legacyKey)
		if err != nil {
			t.Fatal(err)
		}
		return encrypted
	}
	type upgradeRecord struct {
		id, status, selected string
		fields               map[string]any
		auth                 map[string]any
	}
	var fixtures []upgradeRecord
	for _, scenario := range []struct {
		name, beforeStatus, afterStatus string
		enabled, unresolved             bool
	}{
		{"grant", "configured", "configured", true, false},
		{"interrupted-sync", "syncing", "configured", true, false},
		{"disabled-dialog", "syncing", "disabled", false, false},
		{"unresolved", "needs_reauth", "needs_reauth", true, true},
	} {
		owner := core.NewRecord(users)
		owner.Load(map[string]any{"username": "upgrade-" + scenario.name, "email": "upgrade-" + scenario.name + "@example.test", "password": "upgrade-password"})
		if err := app.Save(owner); err != nil {
			t.Fatal(err)
		}
		auth := map[string]any{
			"clientId": "legacy-client", "clientSecret": encrypt("legacy-client-secret"),
			"accessToken": encrypt("legacy-access-token"), "refreshToken": encrypt("legacy-refresh-token"),
			"oauthContext": context, "tokenType": "Bearer", "scope": "read_all,activity:read_all", "expiresAt": "2099-01-01T00:00:00Z", "providerUserId": "legacy-account",
			"oauthState": encrypt("legacy-pending-state"), "oauthCodeVerifier": encrypt("legacy-pending-verifier"), "oauthRedirectURI": "https://upgrade.example.test/settings/plugins/oauth/callback",
		}
		selected := context
		if scenario.unresolved {
			auth["oauthContext"] = nil
			selected = ""
		}
		record := core.NewRecord(instances)
		record.Load(map[string]any{
			"user": owner.Id, "plugin_id": manifest.ID, "enabled": scenario.enabled, "status": scenario.beforeStatus,
			"auth": auth, "config": map[string]any{"plugin": map[string]any{"after": "2021-02-03"}, "host": map[string]any{"privacy": "private", "planned": true, "completed": true, "merge": false}},
			"state":        map[string]any{"cursor": 17, "nested": map[string]any{"preserve": true}},
			"last_error":   map[string]any{"code": "auth_failed", "message": "Historical F diagnosis", "custom": map[string]any{"preserve": true}},
			"last_sync_at": "2020-01-02 03:04:05.000Z", "retry_not_before": "2020-02-03 04:05:06.000Z",
		})
		if err := app.Save(record); err != nil {
			t.Fatal(err)
		}
		// Historical created/updated values make an accidental Save hook or
		// autodate rewrite observable, without relying on elapsed wall-clock time.
		if _, err := app.DB().Update(instances.Name, dbx.Params{"created": "2019-01-02 03:04:05.000Z", "updated": "2019-02-03 04:05:06.000Z"}, dbx.HashExp{"id": record.Id}).Execute(); err != nil {
			t.Fatal(err)
		}
		record, err = app.FindRecordById(instances.Id, record.Id)
		if err != nil {
			t.Fatal(err)
		}
		expectedAuth := pluginsystem.JSONMapFromRecord(record, "auth")
		for _, key := range []string{"oauthState", "oauthCodeVerifier", "oauthRedirectURI"} {
			delete(expectedAuth, key)
		}
		fixtures = append(fixtures, upgradeRecord{record.Id, scenario.afterStatus, selected, record.FieldsData(), expectedAuth})
	}
	// A wrong or absent encryption key during upgrade must not matter: all
	// legacy credential bytes are opaque, and migration has no provider I/O.
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", "")
	if err := addPluginConnectionFields1791330000(app); err != nil {
		t.Fatal(err)
	}
	if err := activatePluginConnections1791330001(app); err != nil {
		t.Fatal(err)
	}
	upgraded, err := app.FindCollectionByNameOrId(instances.Id)
	if err != nil {
		t.Fatal(err)
	}
	if upgraded.CreateRule != nil || upgraded.UpdateRule != nil {
		t.Fatal("upgrade did not lock direct collection writes")
	}
	if len(upgraded.Fields) != len(oldFields)+3 || !reflect.DeepEqual(upgraded.Indexes, oldIndexes) {
		t.Fatal("upgrade changed released field count or unique indexes")
	}
	for name, before := range oldFields {
		after, err := json.Marshal(upgraded.Fields.GetByName(name))
		if err != nil {
			t.Fatal(err)
		}
		if string(before) != string(after) {
			t.Errorf("upgrade changed released field definition %s", name)
		}
	}
	for name, after := range map[string]*string{"list": upgraded.ListRule, "view": upgraded.ViewRule, "delete": upgraded.DeleteRule} {
		if !reflect.DeepEqual(after, oldRules[name]) {
			t.Errorf("upgrade changed %s rule", name)
		}
	}
	if field, ok := upgraded.Fields.GetByName("oauth_token_revision").(*core.TextField); !ok || !field.Hidden {
		t.Fatal("token revision is not hidden text")
	}
	installedAfter, err := app.FindCollectionByNameOrId(installed.Id)
	if err != nil {
		t.Fatal(err)
	}
	installedJSON, err := json.Marshal(installedAfter)
	if err != nil {
		t.Fatal(err)
	}
	if string(installedJSON) != string(installedBefore) {
		t.Fatal("upgrade changed installed_plugins collection")
	}
	cachedPlugin, err := app.FindRecordById(installed.Id, plugin.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cachedPlugin.FieldsData(), pluginBefore) {
		t.Fatal("upgrade changed cached manifest/config/plugin data")
	}
	first := map[string]map[string]any{}
	for _, fixture := range fixtures {
		record, err := app.FindRecordById(upgraded.Id, fixture.id)
		if err != nil {
			t.Fatal(err)
		}
		if record.GetString("status") != fixture.status || record.GetString("auth_context") != fixture.selected {
			t.Errorf("upgraded status/selection %s: %q/%q", fixture.id, record.GetString("status"), record.GetString("auth_context"))
		}
		if record.GetString("connection_generation") == "" || record.GetString("oauth_token_revision") == "" {
			t.Error("upgrade omitted G/T initialization")
		}
		if !reflect.DeepEqual(pluginsystem.JSONMapFromRecord(record, "auth"), fixture.auth) {
			t.Error("upgrade changed F grant/credential metadata beyond pending flow removal")
		}
		for name, before := range fixture.fields {
			if name == "auth" || name == "status" {
				continue
			}
			if !reflect.DeepEqual(record.Get(name), before) {
				t.Errorf("upgrade changed preserved released field %s", name)
			}
		}
		if fixture.selected == "" {
			// A later grant metadata change must never trigger another legacy
			// selection migration once G has initialized the B1 instance.
			auth := pluginsystem.JSONMapFromRecord(record, "auth")
			auth["oauthContext"] = context
			authJSON, err := json.Marshal(auth)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := app.DB().Update(upgraded.Name, dbx.Params{"auth": string(authJSON)}, dbx.HashExp{"id": fixture.id}).Execute(); err != nil {
				t.Fatal(err)
			}
			record, err = app.FindRecordById(upgraded.Id, fixture.id)
			if err != nil {
				t.Fatal(err)
			}
		}
		first[fixture.id] = record.FieldsData()
	}
	if err := addPluginConnectionFields1791330000(app); err != nil {
		t.Fatal(err)
	}
	if err := activatePluginConnections1791330001(app); err != nil {
		t.Fatal(err)
	}
	for id, before := range first {
		record, err := app.FindRecordById(upgraded.Id, id)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(record.FieldsData(), before) {
			t.Errorf("repeated real-schema upgrade changed initialized instance %s", id)
		}
	}
}
