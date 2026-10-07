package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/pocketbase/pocketbase/tools/subscriptions"
	"github.com/pocketbase/pocketbase/tools/types"
	"pocketbase/hooks"
	"pocketbase/pluginsystem"
)

type connectionHostFixture struct {
	app        core.App
	handler    http.Handler
	token      string
	owner      *core.Record
	manifest   pluginsystem.Manifest
	installed  *core.Record
	collection *core.Collection
}

func newConnectionHostFixture(t *testing.T) *connectionHostFixture {
	t.Helper()
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", oauthMetadataTestKey)
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	users := core.NewAuthCollection("connection_users")
	if err := app.Save(users); err != nil {
		t.Fatal(err)
	}
	owner := core.NewRecord(users)
	owner.SetEmail("connection@example.test")
	owner.SetPassword("test-password")
	if err := app.Save(owner); err != nil {
		t.Fatal(err)
	}
	token, err := owner.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	manifest := pluginsystem.Manifest{ManifestVersion: "1.0", ID: "connection-test", Type: pluginsystem.PluginTypeTrails, Name: "Connection test", Version: "1.0", Runtime: pluginsystem.RuntimeManifest{Type: pluginsystem.RuntimeWASM, Entrypoint: "plugin.wasm"}, Capabilities: []pluginsystem.CapabilityManifest{{Name: "list", Version: "1.0", Export: "list"}}, Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{"api": {Type: pluginsystem.AuthTypeAPIKey, Fields: []string{"label"}, SecretField: "apiKey"}}}}
	installedCol := core.NewBaseCollection("installed_plugins")
	installedCol.Fields.Add(&core.TextField{Name: "plugin_id"}, &core.JSONField{Name: "manifest"}, &core.JSONField{Name: "config"})
	if err := app.Save(installedCol); err != nil {
		t.Fatal(err)
	}
	installed := core.NewRecord(installedCol)
	installed.Set("plugin_id", manifest.ID)
	installed.Set("manifest", manifest)
	installed.Set("config", map[string]any{})
	if err := app.Save(installed); err != nil {
		t.Fatal(err)
	}
	instances := core.NewBaseCollection("plugin_instances")
	instances.ViewRule = types.Pointer("user = @request.auth.id")
	instances.ListRule = instances.ViewRule
	instances.DeleteRule = instances.ViewRule
	instances.Fields.Add(&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1}, &core.BoolField{Name: "enabled"})
	for _, name := range []string{"plugin_id", "status", "auth_context", "connection_generation", "last_sync_at", "retry_not_before"} {
		instances.Fields.Add(&core.TextField{Name: name})
	}
	instances.Fields.Add(&core.TextField{Name: "oauth_token_revision", Hidden: true})
	for _, name := range []string{"auth", "config", "state", "last_error"} {
		instances.Fields.Add(&core.JSONField{Name: name})
	}
	instances.Indexes = []string{"CREATE UNIQUE INDEX connection_user_plugin ON plugin_instances (user,plugin_id)"}
	if err := app.Save(instances); err != nil {
		t.Fatal(err)
	}
	app.OnRecordCreateRequest("plugin_instances").BindFunc(hooks.DenyPluginInstanceWriteHandler())
	app.OnRecordUpdateRequest("plugin_instances").BindFunc(hooks.DenyPluginInstanceWriteHandler())
	app.OnRecordEnrich("plugin_instances").BindFunc(hooks.EnrichPluginInstanceHandler())
	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	router.POST("/plugins/instances", PluginSystemInstanceCreate)
	router.PATCH("/plugins/instances/{id}/settings", PluginSystemInstanceSettings)
	handler, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	return &connectionHostFixture{app, handler, token, owner, manifest, installed, instances}
}

func (f *connectionHostFixture) request(t *testing.T, method, path string, body any) (int, map[string]any, string) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.Header.Set("Authorization", f.token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.handler.ServeHTTP(w, req)
	var result map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	return w.Code, result, w.Body.String()
}
func (f *connectionHostFixture) create(t *testing.T, auth map[string]any) *core.Record {
	t.Helper()
	status, result, body := f.request(t, "POST", "/plugins/instances", map[string]any{"plugin_id": f.manifest.ID, "auth": auth, "enabled": true})
	if status != 200 {
		t.Fatalf("create %d %s", status, body)
	}
	r, err := f.app.FindRecordById("plugin_instances", result["id"].(string))
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func (f *connectionHostFixture) patch(t *testing.T, r *core.Record, body map[string]any) (int, map[string]any, string) {
	return f.request(t, "PATCH", "/plugins/instances/"+r.Id+"/settings", body)
}
func connectionResponseCode(result map[string]any) string {
	data, _ := result["data"].(map[string]any)
	operation, _ := data["operation"].(map[string]any)
	code, _ := operation["code"].(string)
	return code
}

func TestHostConnectionCreateOwnerUniqueAndDirectWriteDenial(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "explicit-secret", "label": "public"})
	if r.GetString("user") != f.owner.Id || r.GetString("auth_context") != "api" || r.GetString("connection_generation") == "" || r.GetString("oauth_token_revision") == "" {
		t.Fatal("create omitted owner selection or revisions")
	}
	auth := pluginsystem.JSONMapFromRecord(r, "auth")
	decoded, err := security.Decrypt(auth["apiKey"].(string), oauthMetadataTestKey)
	if err != nil || string(decoded) != "explicit-secret" {
		t.Fatal("secret was not encrypted exactly once")
	}
	status, result, body := f.request(t, "POST", "/plugins/instances", map[string]any{"plugin_id": f.manifest.ID})
	if status != 409 || connectionResponseCode(result) != "plugin_instance_exists" {
		t.Fatalf("unique %d %s", status, body)
	}
	for _, method := range []string{"POST", "PATCH"} {
		path := "/api/collections/plugin_instances/records"
		if method == "PATCH" {
			path += "/" + r.Id
		}
		status, _, body = f.request(t, method, path, map[string]any{"auth": map[string]any{"apiKey": "leak"}})
		if status != 403 {
			t.Errorf("PB %s %d %s", method, status, body)
		}
	}
	for _, patch := range []map[string]any{{"status": "configured"}, {"user": f.owner.Id}, {"auth": map[string]any{"oauthState": "secret"}}, {"auth": map[string]any{"label": nil}}, {"authContext": nil}, {"oauth_token_revision": "revision"}} {
		status, _, body = f.patch(t, r, patch)
		if status != 400 {
			t.Errorf("bad DTO accepted %d %s", status, body)
		}
	}
}

func TestHostExplicitSecretRepairAndGenerationGuards(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "first"})
	generation := r.GetString("connection_generation")
	raw := strings.Repeat("abcdef01", 8)
	r.Set("auth", map[string]any{"apiKey": raw})
	if err := f.app.Save(r); err != nil {
		t.Fatal(err)
	}
	for _, patch := range []map[string]any{{"auth": map[string]any{"apiKey": ""}}, {"config": map[string]any{"host": map[string]any{"categoryRemapDismissedAt": "today"}}}} {
		status, _, body := f.patch(t, r, patch)
		if status != 200 {
			t.Fatalf("placeholder %d %s", status, body)
		}
		stored, _ := f.app.FindRecordById("plugin_instances", r.Id)
		if pluginsystem.JSONMapFromRecord(stored, "auth")["apiKey"] != raw || stored.GetString("connection_generation") != generation {
			t.Fatal("placeholder changed stored bytes or generation")
		}
	}
	status, result, body := f.patch(t, r, map[string]any{"auth": map[string]any{"apiKey": raw}})
	if status != 400 || connectionResponseCode(result) != "connection_generation_required" {
		t.Fatalf("guard required %d %s", status, body)
	}
	status, _, body = f.patch(t, r, map[string]any{"auth": map[string]any{"apiKey": raw}, "expected_connection_generation": generation})
	if status != 200 {
		t.Fatalf("repair %d %s", status, body)
	}
	stored, _ := f.app.FindRecordById("plugin_instances", r.Id)
	value := pluginsystem.JSONMapFromRecord(stored, "auth")["apiKey"].(string)
	decoded, err := security.Decrypt(value, oauthMetadataTestKey)
	if err != nil || string(decoded) != raw {
		t.Fatal("repair was not encrypted exactly once")
	}
	if stored.GetString("connection_generation") == generation {
		t.Fatal("unreadable replacement retained generation")
	}
	current := stored.GetString("connection_generation")
	status, result, body = f.patch(t, r, map[string]any{"auth": map[string]any{"apiKey": raw}, "expected_connection_generation": generation})
	if status != 409 || connectionResponseCode(result) != "plugin_connection_conflict" {
		t.Fatalf("stale same input %d %s", status, body)
	}
	status, _, body = f.patch(t, r, map[string]any{"auth": map[string]any{"apiKey": raw}, "expected_connection_generation": current})
	if status != 200 {
		t.Fatalf("same logical secret %d %s", status, body)
	}
	stored, _ = f.app.FindRecordById("plugin_instances", r.Id)
	if stored.GetString("connection_generation") != current {
		t.Fatal("same logical secret renewed generation")
	}
	status, _, body = f.patch(t, r, map[string]any{"auth": map[string]any{"apiKey": nil}, "expected_connection_generation": current})
	if status != 200 {
		t.Fatalf("null %d %s", status, body)
	}
	stored, _ = f.app.FindRecordById("plugin_instances", r.Id)
	if _, exists := pluginsystem.JSONMapFromRecord(stored, "auth")["apiKey"]; exists {
		t.Fatal("null secret not removed")
	}
}

func TestHostConfigValidationUsesFreshPatchComparison(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "first"})
	manifest := f.manifest
	manifest.ConfigSchema = []pluginsystem.ConfigField{{Key: "endpoint", Type: "string", Required: true}}
	f.installed.Set("manifest", manifest)
	if err := f.app.Save(f.installed); err != nil {
		t.Fatal(err)
	}
	for _, patch := range []map[string]any{{"enabled": false}, {"config": map[string]any{}}, {"config": map[string]any{"host": map[string]any{"categoryRemapDismissedAt": "today"}}}, {"auth": map[string]any{"apiKey": "repair"}, "expected_connection_generation": r.GetString("connection_generation")}} {
		status, _, body := f.patch(t, r, patch)
		if status != 200 {
			t.Fatalf("permitted repair %d %s", status, body)
		}
	}
	before, _ := f.app.FindRecordById("plugin_instances", r.Id)
	status, _, body := f.patch(t, r, map[string]any{"enabled": true})
	if status != 400 {
		t.Fatalf("invalid activation %d %s", status, body)
	}
	status, _, body = f.patch(t, r, map[string]any{"enabled": false, "config": map[string]any{"plugin": map[string]any{"other": "changed"}}})
	if status != 400 {
		t.Fatalf("mixed invalid %d %s", status, body)
	}
	after, _ := f.app.FindRecordById("plugin_instances", r.Id)
	if !reflect.DeepEqual(before.FieldsData(), after.FieldsData()) {
		t.Fatal("invalid mixed patch partially committed")
	}
	status, _, body = f.patch(t, r, map[string]any{"config": map[string]any{"plugin": map[string]any{"endpoint": "valid"}}, "enabled": true})
	if status != 200 {
		t.Fatalf("valid repair %d %s", status, body)
	}
}

func TestHostSettingsFreshCommitPreservesConcurrentAuthAndStatus(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "first"})
	entered, resume := make(chan struct{}), make(chan struct{})
	pluginInstanceSettingsBeforeCommit = func() { close(entered); <-resume }
	t.Cleanup(func() { pluginInstanceSettingsBeforeCommit = nil })
	done := make(chan int, 1)
	go func() {
		status, _, _ := f.patch(t, r, map[string]any{"config": map[string]any{"plugin": map[string]any{"after": "draft"}}})
		done <- status
	}()
	<-entered
	fresh, _ := f.app.FindRecordById("plugin_instances", r.Id)
	auth := pluginsystem.JSONMapFromRecord(fresh, "auth")
	auth["oauthFlowContext"] = "future"
	auth["oauthActivateOnCallback"] = true
	fresh.Set("auth", auth)
	fresh.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
	fresh.Set("status", "rate_limited")
	fresh.Set("last_error", map[string]any{"code": "rate_limited"})
	if err := f.app.Save(fresh); err != nil {
		t.Fatal(err)
	}
	close(resume)
	if status := <-done; status != 200 {
		t.Fatalf("settings %d", status)
	}
	stored, _ := f.app.FindRecordById("plugin_instances", r.Id)
	if stored.GetString("oauth_token_revision") != fresh.GetString("oauth_token_revision") || stored.GetString("status") != "rate_limited" || pluginsystem.JSONMapFromRecord(stored, "auth")["oauthFlowContext"] != "future" {
		t.Fatal("settings overwrote concurrent token flow or status")
	}
	pluginInstanceSettingsBeforeCommit = nil
	status, _, body := f.patch(t, r, map[string]any{"enabled": false})
	if status != 200 {
		t.Fatalf("disable %d %s", status, body)
	}
	stored, _ = f.app.FindRecordById("plugin_instances", r.Id)
	if pluginsystem.JSONMapFromRecord(stored, "auth")["oauthActivateOnCallback"] != false || stored.GetString("status") != "disabled" {
		t.Fatal("disable retained activation intent")
	}
}

func TestCommonConnectionProjectionPreservesWorkingRecordAndSecretPresence(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "first", "label": "public"})
	auth := pluginsystem.JSONMapFromRecord(r, "auth")
	auth["accessToken"] = ""
	auth["refreshToken"] = nil
	auth["oauthState"] = "FLOW-SECRET"
	auth["oauthFlowContext"] = "api"
	r.Set("auth", auth)
	r.Set("state", map[string]any{"nested": map[string]any{"oauthTokenRevision": "T-SECRET", "oauthState": "FLOW-SECRET"}})
	r.Set("last_error", map[string]any{"connectionGeneration": r.GetString("connection_generation"), "oauthTokenRevision": "T-SECRET"})
	r.Set("unknownInternal", "CUSTOM-SECRET")
	r.WithCustomData(true)
	before := r.GetString("auth")
	public, err := pluginsystem.PublicConnectionRecord(f.app, r)
	if err != nil {
		t.Fatal(err)
	}
	if r.GetString("auth") != before {
		t.Fatal("read projection mutated working auth")
	}
	encoded, _ := json.Marshal(public)
	for _, secret := range []string{"FLOW-SECRET", "T-SECRET", "CUSTOM-SECRET", "oauth_token_revision", "oauthState", "oauthFlowContext"} {
		if bytes.Contains(encoded, []byte(secret)) {
			t.Fatalf("public leaked %s: %s", secret, encoded)
		}
	}
	output := pluginsystem.JSONMapFromRecord(public, "auth")
	if output["apiKey"] != "" || output["label"] != "public" {
		t.Fatal("placeholder or public input missing")
	}
	for _, key := range []string{"accessToken", "refreshToken"} {
		if _, exists := output[key]; exists {
			t.Fatalf("absent secret created placeholder %s", key)
		}
	}
	release, ok := pluginsystem.TryLockInstance(r.Id)
	if !ok {
		t.Fatal("lock")
	}
	public, _ = pluginsystem.PublicConnectionRecord(f.app, r)
	if public.PublicExport()["busy"] != true {
		t.Fatal("busy not projected")
	}
	release()
	public, _ = pluginsystem.PublicConnectionRecord(f.app, r)
	if public.PublicExport()["busy"] != false {
		t.Fatal("busy remained after release")
	}
	if err := f.app.Save(r); err != nil {
		t.Fatal(err)
	}
	status, _, body := f.request(t, "GET", "/api/collections/plugin_instances/records/"+r.Id, nil)
	if status != 200 || strings.Contains(body, "FLOW-SECRET") || strings.Contains(body, "T-SECRET") {
		t.Fatalf("common PB view %d %s", status, body)
	}
	stored, _ := f.app.FindRecordById("plugin_instances", r.Id)
	if pluginsystem.JSONMapFromRecord(stored, "auth")["oauthState"] != "FLOW-SECRET" {
		t.Fatal("PB view changed database")
	}
}

func TestConnectionReadAndWriteBoundariesIncludeSuperuserBatchAndExpand(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "API-SECRET"})
	auth := pluginsystem.JSONMapFromRecord(r, "auth")
	auth["oauthState"] = "FLOW-SECRET"
	r.Set("auth", auth)
	r.Set("last_error", map[string]any{"nested": map[string]any{"oauthTokenRevision": "TOKEN-REVISION-SECRET"}})
	if err := f.app.Save(r); err != nil {
		t.Fatal(err)
	}
	superusers, err := f.app.FindCollectionByNameOrId("_superusers")
	if err != nil {
		t.Fatal(err)
	}
	admin := core.NewRecord(superusers)
	admin.SetEmail("admin@example.test")
	admin.SetPassword("admin-password")
	if err := f.app.Save(admin); err != nil {
		t.Fatal(err)
	}
	adminToken, err := admin.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	wrapperCol := core.NewBaseCollection("connection_wrappers")
	wrapperCol.ViewRule = types.Pointer("")
	wrapperCol.Fields.Add(&core.RelationField{Name: "instance", CollectionId: f.collection.Id, MaxSelect: 1})
	if err := f.app.Save(wrapperCol); err != nil {
		t.Fatal(err)
	}
	wrapper := core.NewRecord(wrapperCol)
	wrapper.Set("instance", r.Id)
	if err := f.app.Save(wrapper); err != nil {
		t.Fatal(err)
	}
	f.app.Settings().Batch.Enabled = true
	for _, token := range []string{f.token, adminToken} {
		f.token = token
		for _, path := range []string{"/api/collections/plugin_instances/records/" + r.Id, "/api/collections/plugin_instances/records", "/api/collections/connection_wrappers/records/" + wrapper.Id + "?expand=instance"} {
			status, _, body := f.request(t, "GET", path, nil)
			if status != 200 {
				t.Fatalf("read %d %s", status, body)
			}
			for _, secret := range []string{"FLOW-SECRET", "TOKEN-REVISION-SECRET", "API-SECRET", r.GetString("oauth_token_revision")} {
				if strings.Contains(body, secret) {
					t.Errorf("read leaked %s", secret)
				}
			}
			if !strings.Contains(body, r.GetString("connection_generation")) {
				t.Fatal("read lost public generation")
			}
		}
		for _, method := range []string{"POST", "PATCH"} {
			path := "/api/collections/plugin_instances/records"
			if method == "PATCH" {
				path += "/" + r.Id
			}
			status, _, body := f.request(t, method, path, map[string]any{"user": f.owner.Id, "plugin_id": f.manifest.ID, "status": "configured"})
			if status != 403 {
				t.Errorf("direct %s %d %s", method, status, body)
			}
			status, _, body = f.request(t, "POST", "/api/batch", map[string]any{"requests": []any{map[string]any{"method": method, "url": path, "body": map[string]any{"user": f.owner.Id, "plugin_id": f.manifest.ID, "status": "configured"}}}})
			if status != 400 || !strings.Contains(body, "\"status\":403") {
				t.Errorf("batch %s %d %s", method, status, body)
			}
		}
	}
	status, _, body := f.request(t, "PATCH", "/plugins/instances/"+r.Id+"/settings", map[string]any{"enabled": false})
	if status != 403 {
		t.Fatalf("superuser acted as owner %d %s", status, body)
	}
	stored, _ := f.app.FindRecordById("plugin_instances", r.Id)
	if !stored.GetBool("enabled") || pluginsystem.JSONMapFromRecord(stored, "auth")["oauthState"] != "FLOW-SECRET" {
		t.Fatal("read or denied write changed DB")
	}
}

func TestConnectionSettingsPreservesMalformedAuthForControlPatches(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "secret"})
	r.Set("auth", []any{"unreadable", "preserve"})
	if err := f.app.Save(r); err != nil {
		t.Fatal(err)
	}
	before := r.GetString("auth")
	for _, patch := range []map[string]any{{"enabled": false}, {"config": map[string]any{"host": map[string]any{"categoryRemapDismissedAt": "today"}}}} {
		status, _, body := f.patch(t, r, patch)
		if status != 200 {
			t.Fatalf("control %d %s", status, body)
		}
		stored, _ := f.app.FindRecordById("plugin_instances", r.Id)
		if stored.GetString("auth") != before {
			t.Fatal("control patch replaced malformed auth")
		}
	}
}

func TestRealtimeConnectionProjectionUsesCommonEnrich(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "secret"})
	client := subscriptions.NewDefaultClient()
	client.Set(apis.RealtimeClientAuthKey, f.owner)
	client.Subscribe("plugin_instances/" + r.Id)
	f.app.SubscriptionsBroker().Register(client)
	defer f.app.SubscriptionsBroker().Unregister(client.Id())
	got := make(chan subscriptions.Message, 1)
	go func() {
		select {
		case message := <-client.Channel():
			got <- message
		case <-time.After(5 * time.Second):
		}
	}()
	auth := pluginsystem.JSONMapFromRecord(r, "auth")
	auth["oauthState"] = "REALTIME-FLOW-SECRET"
	r.Set("auth", auth)
	r.Set("state", map[string]any{"oauthTokenRevision": "REALTIME-T-SECRET"})
	if err := f.app.Save(r); err != nil {
		t.Fatal(err)
	}
	select {
	case message := <-got:
		for _, secret := range []string{"REALTIME-FLOW-SECRET", "REALTIME-T-SECRET", r.GetString("oauth_token_revision")} {
			if bytes.Contains(message.Data, []byte(secret)) {
				t.Fatalf("realtime leaked %s: %s", secret, message.Data)
			}
		}
		if !bytes.Contains(message.Data, []byte("\"apiKey\":\"\"")) {
			t.Fatalf("realtime secret placeholder missing: %s", message.Data)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("realtime event not received")
	}
	stored, _ := f.app.FindRecordById("plugin_instances", r.Id)
	if pluginsystem.JSONMapFromRecord(stored, "auth")["oauthState"] != "REALTIME-FLOW-SECRET" {
		t.Fatal("realtime modified DB")
	}
}

func TestInactiveCredentialsAndContextChangeUseConnectionFieldSets(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "active-api"})
	oauth := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeOAuth2, Fields: []string{"clientId", "clientSecret"}, AuthorizationURL: "https://example.test/auth", TokenURL: "https://example.test/token", Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModeHost}}
	f.manifest.Auth.Contexts["oauth"] = oauth
	f.installed.Set("manifest", f.manifest)
	if err := f.app.Save(f.installed); err != nil {
		t.Fatal(err)
	}
	secret, err := security.Encrypt([]byte("old-client-secret"), oauthMetadataTestKey)
	if err != nil {
		t.Fatal(err)
	}
	auth := pluginsystem.JSONMapFromRecord(r, "auth")
	auth["clientId"] = "old-client"
	auth["clientSecret"] = secret
	auth["accessToken"] = "opaque-grant"
	auth["refreshToken"] = "opaque-refresh"
	auth["oauthContext"] = "oauth"
	auth["oauthState"] = "pending"
	auth["oauthFlowContext"] = "oauth"
	r.Set("auth", auth)
	r.Set("state", map[string]any{"cursor": "preserve"})
	r.Set("last_error", map[string]any{"code": "historical"})
	if err := f.app.Save(r); err != nil {
		t.Fatal(err)
	}
	generation, tokenRevision := r.GetString("connection_generation"), r.GetString("oauth_token_revision")
	status, _, body := f.patch(t, r, map[string]any{"auth": map[string]any{"clientSecret": "new-client-secret"}})
	if status != 200 {
		t.Fatalf("inactive credential %d %s", status, body)
	}
	stored, _ := f.app.FindRecordById("plugin_instances", r.Id)
	if stored.GetString("connection_generation") != generation || stored.GetString("oauth_token_revision") == tokenRevision {
		t.Fatal("inactive grant clear changed wrong revision")
	}
	if pluginsystem.JSONMapFromRecord(stored, "state")["cursor"] != "preserve" || pluginsystem.JSONMapFromRecord(stored, "last_error")["code"] != "historical" {
		t.Fatal("inactive credential changed state or diagnosis")
	}
	auth = pluginsystem.JSONMapFromRecord(stored, "auth")
	for _, key := range []string{"accessToken", "refreshToken", "oauthContext", "oauthState"} {
		if _, exists := auth[key]; exists {
			t.Fatalf("inactive grant field not cleared: %s", key)
		}
	}
	auth["accessToken"] = "opaque-grant"
	auth["refreshToken"] = "opaque-refresh"
	auth["oauthContext"] = "oauth"
	auth["oauthState"] = "pending"
	stored.Set("auth", auth)
	if err := f.app.Save(stored); err != nil {
		t.Fatal(err)
	}
	tokenRevision = stored.GetString("oauth_token_revision")
	status, _, body = f.patch(t, r, map[string]any{"authContext": "oauth", "expected_connection_generation": generation})
	if status != 200 {
		t.Fatalf("context switch %d %s", status, body)
	}
	stored, _ = f.app.FindRecordById("plugin_instances", r.Id)
	auth = pluginsystem.JSONMapFromRecord(stored, "auth")
	if stored.GetString("auth_context") != "oauth" || stored.GetString("connection_generation") == generation || stored.GetString("oauth_token_revision") != tokenRevision || auth["accessToken"] != "opaque-grant" || auth["oauthContext"] != "oauth" {
		t.Fatal("context switch rewrote grant or wrong revisions")
	}
	if len(pluginsystem.JSONMapFromRecord(stored, "state")) != 0 || len(pluginsystem.JSONMapFromRecord(stored, "last_error")) != 0 {
		t.Fatal("new generation retained old state/diagnosis")
	}
	if _, exists := auth["oauthState"]; exists {
		t.Fatal("new generation retained pending flow")
	}
}

func TestSharedAccessTokenCredentialRequiresAPISelectionAndConsumesOAuthGrant(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "active-api"})
	f.manifest.Auth.Contexts["token"] = pluginsystem.AuthContext{Type: pluginsystem.AuthTypeBearer, SecretField: "accessToken"}
	f.manifest.Auth.Contexts["oauth"] = pluginsystem.AuthContext{Type: pluginsystem.AuthTypeOAuth2, Fields: []string{"clientId", "clientSecret"}, AuthorizationURL: "https://example.test/auth", TokenURL: "https://example.test/token", Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModeHost}}
	f.installed.Set("manifest", f.manifest)
	if err := f.app.Save(f.installed); err != nil {
		t.Fatal(err)
	}
	r.Set("auth_context", "oauth")
	r.Set("auth", map[string]any{"clientId": "client", "accessToken": "old-grant", "refreshToken": "old-refresh", "oauthContext": "oauth", "scope": "old-scope"})
	if err := f.app.Save(r); err != nil {
		t.Fatal(err)
	}
	generation, tokenRevision := r.GetString("connection_generation"), r.GetString("oauth_token_revision")
	status, _, body := f.patch(t, r, map[string]any{"auth": map[string]any{"accessToken": "api-token"}, "expected_connection_generation": generation})
	if status != 400 {
		t.Fatalf("OAuth grant edited %d %s", status, body)
	}
	status, _, body = f.patch(t, r, map[string]any{"authContext": "token", "auth": map[string]any{"accessToken": "api-token", "clientSecret": "new-inactive-client"}, "expected_connection_generation": generation})
	if status != 200 {
		t.Fatalf("API-key switch %d %s", status, body)
	}
	stored, _ := f.app.FindRecordById("plugin_instances", r.Id)
	auth := pluginsystem.JSONMapFromRecord(stored, "auth")
	value, err := security.Decrypt(auth["accessToken"].(string), oauthMetadataTestKey)
	if err != nil || string(value) != "api-token" {
		t.Fatal("explicit token lost or doubly encrypted")
	}
	for _, key := range []string{"oauthContext", "refreshToken", "scope"} {
		if _, exists := auth[key]; exists {
			t.Fatalf("old grant metadata %s remained", key)
		}
	}
	if stored.GetString("oauth_token_revision") == tokenRevision || stored.GetString("connection_generation") == generation {
		t.Fatal("shared-key switch did not renew connection revisions")
	}
}

func TestLegacyOAuthGrantWithoutProvenanceIsClearedWhenClientInputsChange(t *testing.T) {
	for _, input := range []string{"clientId", "clientSecret"} {
		t.Run(input, func(t *testing.T) {
			f := newOAuthMetadataFixture(t, func(auth map[string]any) { delete(auth, "oauthContext") })
			before, _ := f.storedAuth(t)
			f.request(t, http.MethodPatch, "/api/plugins/instances/"+f.instanceID+"/settings", map[string]any{
				"auth":                           map[string]any{input: "replacement-client-input"},
				"expected_connection_generation": before.GetString("connection_generation"),
			})
			fresh, auth := f.storedAuth(t)
			for _, field := range []string{"accessToken", "refreshToken", "oauthContext", "expiresAt", "scope"} {
				if _, exists := auth[field]; exists {
					t.Errorf("changed client input retained grant field %s", field)
				}
			}
			if fresh.GetString("connection_generation") == before.GetString("connection_generation") || fresh.GetString("oauth_token_revision") == before.GetString("oauth_token_revision") {
				t.Fatal("changed client input failed to rotate G/T")
			}
			if auth[input] != "replacement-client-input" || f.tokenRequests.Load() != 0 {
				t.Fatal("repair lost explicit input or contacted provider")
			}
		})
	}
}
