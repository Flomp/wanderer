package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	for _, name := range []string{"plugin_id", "status", "auth_context", "connection_generation"} {
		instances.Fields.Add(&core.TextField{Name: name})
	}
	instances.Fields.Add(&core.DateField{Name: "last_sync_at"}, &core.DateField{Name: "retry_not_before"})
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

// Exercise the router with net/http's real server body, which may return the
// final bytes and EOF in one read. bytes.Reader-only handler tests do not cover
// PocketBase's rewind at that boundary.
func connectionHTTPJSON(t *testing.T, f *connectionHostFixture, server *httptest.Server, method, path, body string) (int, map[string]any, string) {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", f.token)
	req.Header.Set("Content-Type", "application/json")
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return response.StatusCode, result, string(data)
}

func TestHostInstanceWritesThroughRealHTTPRouter(t *testing.T) {
	f := newConnectionHostFixture(t)
	server := httptest.NewServer(f.handler)
	t.Cleanup(server.Close)
	status, result, body := connectionHTTPJSON(t, f, server, http.MethodPost, "/plugins/instances", `{"plugin_id":"connection-test","auth":{"apiKey":"initial"},"enabled":true}`)
	if status != http.StatusOK {
		t.Fatalf("HTTP create %d: %s", status, body)
	}
	id := result["id"].(string)
	path := "/plugins/instances/" + id + "/settings"
	for _, input := range []string{`{"config":{"plugin":{"option":"saved"}}}`, `{"enabled":false}`} {
		status, _, body = connectionHTTPJSON(t, f, server, http.MethodPatch, path, input)
		if status != http.StatusOK {
			t.Fatalf("HTTP settings %d: %s", status, body)
		}
	}
	before, err := f.app.FindRecordById("plugin_instances", id)
	if err != nil {
		t.Fatal(err)
	}
	if before.GetBool("enabled") || before.GetString("status") != "disabled" || pluginsystem.JSONMapFromRecord(before, "config")["plugin"].(map[string]any)["option"] != "saved" {
		t.Fatal("HTTP save/disable did not persist their fields")
	}
	before.Set("auth", map[string]any{"apiKey": "unreadable-legacy-ciphertext"})
	if err := f.app.Save(before); err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(map[string]any{"auth": map[string]any{"apiKey": "repaired"}, "expected_connection_generation": before.GetString("connection_generation")})
	status, _, body = connectionHTTPJSON(t, f, server, http.MethodPatch, path, string(input))
	if status != http.StatusOK {
		t.Fatalf("HTTP repair %d: %s", status, body)
	}
	status, _, body = connectionHTTPJSON(t, f, server, http.MethodPatch, path, `{"enabled":true}`)
	if status != http.StatusOK {
		t.Fatalf("HTTP enable %d: %s", status, body)
	}
	fresh, _ := f.app.FindRecordById("plugin_instances", id)
	decoded, err := security.Decrypt(pluginsystem.JSONMapFromRecord(fresh, "auth")["apiKey"].(string), oauthMetadataTestKey)
	if err != nil || string(decoded) != "repaired" || !fresh.GetBool("enabled") || fresh.GetString("status") != "configured" || fresh.GetString("connection_generation") == before.GetString("connection_generation") {
		t.Fatal("real HTTP repair/enable did not establish the new connection")
	}
}

func TestHostInstanceHTTPBodyRejectsTrailingObjectsAndInvalidJSON(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "initial"})
	server := httptest.NewServer(f.handler)
	t.Cleanup(server.Close)
	path := "/plugins/instances/" + r.Id + "/settings"
	for _, input := range []string{`{"enabled":false}{}`, `{"enabled":false}null`, `[]`, `null`, `{`, `{"auth":null}`} {
		status, _, body := connectionHTTPJSON(t, f, server, http.MethodPatch, path, input)
		if status != http.StatusBadRequest {
			t.Fatalf("invalid JSON %q returned %d: %s", input, status, body)
		}
		fresh, _ := f.app.FindRecordById("plugin_instances", r.Id)
		if !fresh.GetBool("enabled") || fresh.GetString("connection_generation") != r.GetString("connection_generation") {
			t.Fatal("rejected body changed the connection")
		}
	}
	status, _, body := connectionHTTPJSON(t, f, server, http.MethodPatch, path, "  {\"enabled\":false} \n\t")
	if status != http.StatusOK {
		t.Fatalf("one object with whitespace rejected %d: %s", status, body)
	}
}

func TestHostInstanceHTTPBodyHonorsRouterBodyLimit(t *testing.T) {
	f := newConnectionHostFixture(t)
	router, err := apis.NewRouter(f.app)
	if err != nil {
		t.Fatal(err)
	}
	router.POST("/plugins/instances", PluginSystemInstanceCreate).Bind(apis.BodyLimit(128))
	handler, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	input, _ := json.Marshal(map[string]any{"plugin_id": f.manifest.ID, "auth": map[string]any{"apiKey": strings.Repeat("a", 256)}})
	status, _, body := connectionHTTPJSON(t, f, server, http.MethodPost, "/plugins/instances", string(input))
	if status != http.StatusRequestEntityTooLarge {
		t.Fatalf("HTTP body limit returned %d: %s", status, body)
	}
	count, err := f.app.CountRecords("plugin_instances")
	if err != nil || count != 0 {
		t.Fatalf("body-limit rejection wrote an instance: count=%d, error=%v", count, err)
	}
}

func TestHostCreateRejectsUnreadyActivationAndHostOwnedFields(t *testing.T) {
	f := newConnectionHostFixture(t)
	status, _, body := f.request(t, http.MethodPost, "/plugins/instances", map[string]any{"plugin_id": f.manifest.ID, "enabled": true})
	if status != http.StatusBadRequest {
		t.Fatalf("create without required input returned %d: %s", status, body)
	}
	for _, field := range []string{"user", "status", "state", "last_error", "connection_generation", "oauth_token_revision", "auth_context", "last_sync_at", "retry_not_before", "busy", "expected_connection_generation"} {
		status, _, body = f.request(t, http.MethodPost, "/plugins/instances", map[string]any{"plugin_id": f.manifest.ID, field: "client-write"})
		if status != http.StatusBadRequest {
			t.Errorf("create accepted host-owned %s: %d %s", field, status, body)
		}
	}
	count, err := f.app.CountRecords("plugin_instances")
	if err != nil || count != 0 {
		t.Fatalf("rejected create wrote an instance: count=%d, error=%v", count, err)
	}
	status, result, body := f.request(t, http.MethodPost, "/plugins/instances", map[string]any{"plugin_id": f.manifest.ID, "auth": map[string]any{"apiKey": "saved-credential"}, "enabled": false})
	if status != http.StatusOK {
		t.Fatalf("disabled create returned %d: %s", status, body)
	}
	fresh, _ := f.app.FindRecordById("plugin_instances", result["id"].(string))
	if fresh.GetString("user") != f.owner.Id || fresh.GetString("plugin_id") != f.manifest.ID || fresh.GetBool("enabled") || fresh.GetString("status") != "disabled" || fresh.GetString("auth_context") != "api" {
		t.Fatal("create did not derive owner/selection/disabled readiness")
	}
	if fresh.GetString("connection_generation") == "" || fresh.GetString("oauth_token_revision") == "" || len(pluginsystem.JSONMapFromRecord(fresh, "state")) != 0 || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || fresh.GetString("last_sync_at") != "" || fresh.GetString("retry_not_before") != "" {
		t.Fatal("create omitted revisions or invented sync state")
	}
	decoded, err := security.Decrypt(pluginsystem.JSONMapFromRecord(fresh, "auth")["apiKey"].(string), oauthMetadataTestKey)
	if err != nil || string(decoded) != "saved-credential" {
		t.Fatal("disabled create did not encrypt the explicit input")
	}
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
	release, acquired := pluginsystem.TryLockInstance(r.Id)
	if !acquired {
		t.Fatal("lock not acquired")
	}
	defer release()
	for index, token := range []string{f.token, adminToken} {
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
			if index == 0 && !strings.Contains(body, `"busy":true`) || index == 1 && strings.Contains(body, `"busy"`) {
				t.Fatalf("incorrect owner-only busy projection: %s", body)
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

func fixtureWithOAuthGrantAndAlternativeContexts(t *testing.T) (*connectionHostFixture, *core.Record) {
	t.Helper()
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "separate-api-input"})
	oauth := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeOAuth2, Fields: []string{"clientId", "clientSecret"}, AuthorizationURL: "https://example.test/auth", TokenURL: "https://example.test/token", Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModeHost}}
	f.manifest.Auth.Contexts["oauth"] = oauth
	f.manifest.Auth.Contexts["other-oauth"] = oauth
	f.manifest.Auth.Contexts["bearer"] = pluginsystem.AuthContext{Type: pluginsystem.AuthTypeBearer, SecretField: "accessToken"}
	f.manifest.Auth.Contexts["shared-api"] = pluginsystem.AuthContext{Type: pluginsystem.AuthTypeAPIKey, SecretField: "accessToken"}
	f.installed.Set("manifest", f.manifest)
	if err := f.app.Save(f.installed); err != nil {
		t.Fatal(err)
	}
	auth := pluginsystem.JSONMapFromRecord(r, "auth")
	for key, value := range map[string]any{"clientId": "client", "clientSecret": "stored-client", "accessToken": "oauth-grant", "refreshToken": "oauth-refresh", "oauthContext": "oauth", "scope": "read", "expiresAt": time.Now().Add(time.Hour).Unix(), "providerUserId": "old-account", "oauthState": "pending-state", "oauthFlowContext": "oauth"} {
		auth[key] = value
	}
	r.Set("auth_context", "oauth")
	r.Set("auth", auth)
	r.Set("state", map[string]any{"cursor": "old-account-cursor"})
	r.Set("last_error", map[string]any{"code": "historical"})
	r.Set("retry_not_before", "2030-01-01 00:00:00.000Z")
	if err := f.app.Save(r); err != nil {
		t.Fatal(err)
	}
	return f, r
}

func TestHostContextOnlySwitchCannotReuseOAuthGrant(t *testing.T) {
	for _, target := range []string{"bearer", "shared-api", "api", "other-oauth"} {
		t.Run(target, func(t *testing.T) {
			f, before := fixtureWithOAuthGrantAndAlternativeContexts(t)
			status, _, body := f.patch(t, before, map[string]any{"authContext": target, "expected_connection_generation": before.GetString("connection_generation")})
			if status != http.StatusOK {
				t.Fatalf("context switch %d: %s", status, body)
			}
			fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
			auth := pluginsystem.JSONMapFromRecord(fresh, "auth")
			for _, field := range append(append([]string{"accessToken", "refreshToken"}, pluginsystem.InternalOAuthMetadataFields()...), pluginsystem.InternalOAuthTransientFields()...) {
				if _, exists := auth[field]; exists {
					t.Errorf("foreign OAuth grant field %q survived selection change", field)
				}
			}
			if fresh.GetString("auth_context") != target || fresh.GetString("connection_generation") == before.GetString("connection_generation") || fresh.GetString("oauth_token_revision") == before.GetString("oauth_token_revision") {
				t.Fatal("selection switch failed to renew G/T")
			}
			if len(pluginsystem.JSONMapFromRecord(fresh, "state")) != 0 || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || fresh.GetString("retry_not_before") != "" {
				t.Fatal("new selection retained connection state or backoff")
			}
			if auth["clientId"] != "client" || auth["clientSecret"] != "stored-client" || auth["apiKey"] != pluginsystem.JSONMapFromRecord(before, "auth")["apiKey"] {
				t.Fatal("grant clear damaged separate credential inputs")
			}
			wantStatus := "needs_auth"
			if target == "api" {
				wantStatus = "configured" // This context has its own stored API key.
			} else {
				wantStatus = "disabled"
				if fresh.GetBool("enabled") {
					t.Fatal("selection without its own required credentials remained enabled")
				}
			}
			if fresh.GetString("status") != wantStatus {
				t.Fatalf("status = %q, want %q", fresh.GetString("status"), wantStatus)
			}
		})
	}
}

func TestHostOAuthGrantClearDisablesConnectionOrRejectsExplicitActivation(t *testing.T) {
	for _, input := range []string{"clientId", "clientSecret", "context"} {
		t.Run(input, func(t *testing.T) {
			f, before := fixtureWithOAuthGrantAndAlternativeContexts(t)
			patch := map[string]any{"expected_connection_generation": before.GetString("connection_generation"), "enabled": true}
			if input == "context" {
				patch["authContext"] = "other-oauth"
			} else {
				patch["auth"] = map[string]any{input: "replacement"}
			}
			status, _, body := f.patch(t, before, patch)
			if status != http.StatusBadRequest {
				t.Fatalf("grant clear + explicit activation %d: %s", status, body)
			}
			fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
			assertConnectionRecordFieldSet(t, before, fresh, nil)
			delete(patch, "enabled")
			status, _, body = f.patch(t, before, patch)
			if status != http.StatusOK {
				t.Fatalf("grant clear %d: %s", status, body)
			}
			fresh, _ = f.app.FindRecordById("plugin_instances", before.Id)
			if fresh.GetBool("enabled") || fresh.GetString("status") != "disabled" || fresh.GetString("connection_generation") == before.GetString("connection_generation") || fresh.GetString("oauth_token_revision") == before.GetString("oauth_token_revision") {
				t.Fatal("grant-clearing active edit did not disable and renew G/T")
			}
			for _, field := range []string{"accessToken", "refreshToken", "oauthContext"} {
				if _, exists := pluginsystem.JSONMapFromRecord(fresh, "auth")[field]; exists {
					t.Fatalf("grant clear retained %s", field)
				}
			}
			if input != "context" {
				value := pluginsystem.JSONMapFromRecord(fresh, "auth")[input].(string)
				if input == "clientSecret" {
					decoded, err := security.Decrypt(value, oauthMetadataTestKey)
					if err != nil {
						t.Fatal(err)
					}
					value = string(decoded)
				}
				if value != "replacement" {
					t.Fatal("grant clear lost the replacement client credential")
				}
			}
		})
	}
}

func TestHostOAuthActivationRejectsForeignGrantWithoutChangingAnyFields(t *testing.T) {
	f, before := fixtureWithOAuthGrantAndAlternativeContexts(t)
	before.Set("auth_context", "other-oauth")
	before.Set("enabled", false)
	before.Set("status", "disabled")
	if err := f.app.Save(before); err != nil {
		t.Fatal(err)
	}
	status, _, body := f.patch(t, before, map[string]any{"enabled": true})
	if status != http.StatusBadRequest {
		t.Fatalf("foreign grant activation %d: %s", status, body)
	}
	fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
	assertConnectionRecordFieldSet(t, before, fresh, nil)
}

func assertConnectionRecordFieldSet(t *testing.T, before, fresh *core.Record, changed map[string]bool) {
	t.Helper()
	for _, field := range before.Collection().Fields {
		name := field.GetName()
		if name == "updated" || changed[name] {
			continue
		}
		if !reflect.DeepEqual(before.Get(name), fresh.Get(name)) {
			t.Errorf("unexpected write to %s: before=%v, after=%v", name, before.Get(name), fresh.Get(name))
		}
	}
}

func TestHostResolvesMissingSoleSelectionWithoutResettingConnection(t *testing.T) {
	for _, oauth := range []bool{false, true} {
		t.Run(map[bool]string{false: "api", true: "oauth"}[oauth], func(t *testing.T) {
			f := newConnectionHostFixture(t)
			before := f.create(t, map[string]any{"apiKey": "stored-input"})
			selection := "api"
			if oauth {
				selection = "oauth"
				f.manifest.Auth.Contexts = map[string]pluginsystem.AuthContext{"oauth": {Type: pluginsystem.AuthTypeOAuth2, Fields: []string{"clientId", "clientSecret"}, AuthorizationURL: "https://example.test/auth", TokenURL: "https://example.test/token", Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModeHost}}}
				f.installed.Set("manifest", f.manifest)
				if err := f.app.Save(f.installed); err != nil {
					t.Fatal(err)
				}
				before.Set("auth", map[string]any{"clientId": "legacy-client", "clientSecret": "legacy-secret", "accessToken": "legacy-grant", "refreshToken": "legacy-refresh", "expiresAt": time.Now().Add(time.Hour).Unix()})
			}
			before.Set("auth_context", "")
			before.Set("state", map[string]any{"cursor": "preserve"})
			before.Set("status", "rate_limited")
			before.Set("last_error", map[string]any{"code": "rate_limited", "connectionGeneration": before.GetString("connection_generation")})
			before.Set("retry_not_before", "2030-01-01 00:00:00.000Z")
			if err := f.app.Save(before); err != nil {
				t.Fatal(err)
			}
			status, result, body := f.patch(t, before, map[string]any{"authContext": selection})
			if status != http.StatusBadRequest || connectionResponseCode(result) != "connection_generation_required" {
				t.Fatalf("explicit selection without guard %d: %s", status, body)
			}
			status, _, body = f.patch(t, before, map[string]any{"authContext": selection, "expected_connection_generation": before.GetString("connection_generation")})
			if status != http.StatusOK {
				t.Fatalf("sole selection resolution %d: %s", status, body)
			}
			fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
			assertConnectionRecordFieldSet(t, before, fresh, map[string]bool{"auth_context": true})
			if fresh.GetString("auth_context") != selection || fresh.GetString("status") != "rate_limited" {
				t.Fatal("sole context selection not resolved")
			}
		})
	}
}

func TestHostSoleSelectionWithChangedCredentialsStartsNewConnection(t *testing.T) {
	f := newConnectionHostFixture(t)
	before := f.create(t, map[string]any{"apiKey": "original"})
	before.Set("auth_context", "")
	before.Set("state", map[string]any{"cursor": "old"})
	if err := f.app.Save(before); err != nil {
		t.Fatal(err)
	}
	status, _, body := f.patch(t, before, map[string]any{"authContext": "api", "auth": map[string]any{"apiKey": "replacement"}, "expected_connection_generation": before.GetString("connection_generation")})
	if status != http.StatusOK {
		t.Fatalf("credential replacement %d: %s", status, body)
	}
	fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
	if fresh.GetString("connection_generation") == before.GetString("connection_generation") || len(pluginsystem.JSONMapFromRecord(fresh, "state")) != 0 {
		t.Fatal("changed active credential treated as mere selection resolution")
	}
}

func TestHostSoleSelectionResolutionConsumesForeignGrantAndResetsConnection(t *testing.T) {
	for _, oauth := range []bool{false, true} {
		t.Run(map[bool]string{false: "api", true: "oauth"}[oauth], func(t *testing.T) {
			f := newConnectionHostFixture(t)
			before := f.create(t, map[string]any{"apiKey": "own-input"})
			selection := "api"
			if oauth {
				selection = "oauth"
				f.manifest.Auth.Contexts = map[string]pluginsystem.AuthContext{"oauth": {Type: pluginsystem.AuthTypeOAuth2, Fields: []string{"clientId", "clientSecret"}, AuthorizationURL: "https://example.test/auth", TokenURL: "https://example.test/token", Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModeHost}}}
				f.installed.Set("manifest", f.manifest)
				if err := f.app.Save(f.installed); err != nil {
					t.Fatal(err)
				}
			}
			auth := pluginsystem.JSONMapFromRecord(before, "auth")
			auth["accessToken"], auth["oauthContext"] = "foreign-grant", "foreign-context"
			before.Set("auth_context", "")
			before.Set("auth", auth)
			before.Set("state", map[string]any{"cursor": "foreign-account"})
			before.Set("last_error", map[string]any{"code": "historical"})
			before.Set("retry_not_before", "2030-01-01 00:00:00.000Z")
			if err := f.app.Save(before); err != nil {
				t.Fatal(err)
			}
			status, _, body := f.patch(t, before, map[string]any{"authContext": selection, "expected_connection_generation": before.GetString("connection_generation")})
			if status != http.StatusOK {
				t.Fatalf("foreign grant repair %d: %s", status, body)
			}
			fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
			if fresh.GetString("connection_generation") == before.GetString("connection_generation") || fresh.GetString("oauth_token_revision") == before.GetString("oauth_token_revision") {
				t.Fatal("foreign grant clear treated as compatible selection resolution")
			}
			for _, field := range []string{"accessToken", "oauthContext"} {
				if _, exists := pluginsystem.JSONMapFromRecord(fresh, "auth")[field]; exists {
					t.Fatalf("foreign grant field %q preserved", field)
				}
			}
			if len(pluginsystem.JSONMapFromRecord(fresh, "state")) != 0 || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || fresh.GetString("retry_not_before") != "" {
				t.Fatal("foreign grant clear retained cursor/diagnosis/backoff")
			}
		})
	}
}

func TestHostSettingsRejectsEveryOAuthInternalCredentialKey(t *testing.T) {
	f, before := fixtureWithOAuthGrantAndAlternativeContexts(t)
	fields := append(append(append([]string{"accessToken", "refreshToken"}, pluginsystem.InternalOAuthMetadataFields()...), pluginsystem.InternalOAuthTransientFields()...), "unexpected")
	for _, field := range fields {
		status, _, body := f.patch(t, before, map[string]any{"auth": map[string]any{field: "client-input"}, "expected_connection_generation": before.GetString("connection_generation")})
		if status != http.StatusBadRequest {
			t.Errorf("OAuth internal field %q accepted: %d %s", field, status, body)
		}
		fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
		assertConnectionRecordFieldSet(t, before, fresh, nil)
	}
}

func TestHostEnableReadinessAndControlWriterFieldSets(t *testing.T) {
	f := newConnectionHostFixture(t)
	before := f.create(t, map[string]any{"apiKey": "working"})
	auth := pluginsystem.JSONMapFromRecord(before, "auth")
	auth["oauthState"] = "pending"
	auth["oauthFlowContext"] = "future"
	auth["oauthActivateOnCallback"] = true
	before.Set("auth", auth)
	before.Set("state", map[string]any{"cursor": "preserve"})
	before.Set("last_error", map[string]any{"code": "rate_limited"})
	before.Set("retry_not_before", "2030-01-01 00:00:00.000Z")
	before.Set("last_sync_at", "2026-01-01 00:00:00.000Z")
	if err := f.app.Save(before); err != nil {
		t.Fatal(err)
	}
	status, _, body := f.patch(t, before, map[string]any{"enabled": false})
	if status != http.StatusOK {
		t.Fatalf("disable %d %s", status, body)
	}
	disabled, _ := f.app.FindRecordById("plugin_instances", before.Id)
	assertConnectionRecordFieldSet(t, before, disabled, map[string]bool{"enabled": true, "status": true, "auth": true})
	expectedAuth := pluginsystem.CloneJSONMap(auth)
	expectedAuth["oauthActivateOnCallback"] = false
	if !reflect.DeepEqual(expectedAuth, pluginsystem.JSONMapFromRecord(disabled, "auth")) {
		t.Fatal("disable changed credentials or flow fields beyond activation intent")
	}
	status, _, body = f.patch(t, disabled, map[string]any{"enabled": true})
	if status != http.StatusOK {
		t.Fatalf("ready activation %d %s", status, body)
	}
	enabled, _ := f.app.FindRecordById("plugin_instances", before.Id)
	assertConnectionRecordFieldSet(t, disabled, enabled, map[string]bool{"enabled": true, "status": true})
	if !enabled.GetBool("enabled") || enabled.GetString("status") != "configured" {
		t.Fatal("ready activation failed")
	}
	status, _, body = f.patch(t, enabled, map[string]any{"enabled": false, "auth": map[string]any{"apiKey": nil}, "expected_connection_generation": enabled.GetString("connection_generation")})
	if status != http.StatusOK {
		t.Fatalf("remove credential %d %s", status, body)
	}
	unready, _ := f.app.FindRecordById("plugin_instances", before.Id)
	status, _, body = f.patch(t, unready, map[string]any{"enabled": true})
	if status != http.StatusBadRequest {
		t.Fatalf("unready activation %d %s", status, body)
	}
	fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
	assertConnectionRecordFieldSet(t, unready, fresh, nil)
}

func TestBusyIsAbsentFromNonOwnerProjection(t *testing.T) {
	f := newConnectionHostFixture(t)
	r := f.create(t, map[string]any{"apiKey": "working"})
	release, acquired := pluginsystem.TryLockInstance(r.Id)
	if !acquired {
		t.Fatal("lock not acquired")
	}
	defer release()
	projected := r.Fresh()
	pluginsystem.ProjectConnectionRecord(f.app, projected, false)
	if _, exists := projected.PublicExport()["busy"]; exists {
		t.Fatal("non-owner projection exposed process busy state")
	}
	owner, _ := pluginsystem.PublicConnectionRecord(f.app, r)
	if owner.PublicExport()["busy"] != true {
		t.Fatal("owner projection omitted held lock")
	}
}

func TestSoleContextRepairOnlyDisprovesCurrentEmptySelectionDiagnosis(t *testing.T) {
	for _, test := range []struct {
		name, code, context, generation string
		omitContext, omitGeneration     bool
		enabled, clear                  bool
	}{
		{name: "matching enabled", code: "invalid_configuration", enabled: true, clear: true},
		{name: "matching disabled", code: "invalid_configuration", clear: true},
		{name: "foreign context", code: "invalid_configuration", context: "other", enabled: true},
		{name: "foreign generation", code: "invalid_configuration", generation: "other-generation", enabled: true},
		{name: "empty generation", code: "invalid_configuration", generation: "empty", enabled: true},
		{name: "missing context", code: "invalid_configuration", omitContext: true, enabled: true},
		{name: "missing generation", code: "invalid_configuration", omitGeneration: true, enabled: true},
		{name: "different error", code: "provider_unavailable", enabled: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newConnectionHostFixture(t)
			before := f.create(t, map[string]any{"apiKey": "stored-input", "label": "stored-public"})
			before.Set("auth_context", "")
			before.Set("enabled", test.enabled)
			before.Set("status", "error")
			before.Set("state", map[string]any{"cursor": "preserve"})
			before.Set("retry_not_before", "2030-01-01 00:00:00.000Z")
			before.Set("last_sync_at", "2026-01-01 00:00:00.000Z")
			auth := pluginsystem.JSONMapFromRecord(before, "auth")
			auth["oauthState"], auth["oauthFlowContext"], auth["oauthActivateOnCallback"] = "pending-state", "pending-context", true
			before.Set("auth", auth)
			diagnosis := map[string]any{"code": test.code, "operation": pluginsystem.OperationLogin}
			if !test.omitContext {
				diagnosis["authContext"] = test.context
			}
			if !test.omitGeneration {
				generation := before.GetString("connection_generation")
				if test.generation == "empty" {
					generation = ""
				} else if test.generation != "" {
					generation = test.generation
				}
				diagnosis["connectionGeneration"] = generation
			}
			before.Set("last_error", diagnosis)
			if err := f.app.Save(before); err != nil {
				t.Fatal(err)
			}
			status, _, body := f.patch(t, before, map[string]any{"authContext": "api", "expected_connection_generation": before.GetString("connection_generation")})
			if status != http.StatusOK {
				t.Fatalf("sole repair %d: %s", status, body)
			}
			fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
			allowed := map[string]bool{"auth_context": true}
			if test.clear {
				allowed["status"], allowed["last_error"] = true, true
				wantStatus := "disabled"
				if test.enabled {
					wantStatus = "configured"
				}
				if len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || fresh.GetString("status") != wantStatus {
					t.Fatal("successful selection repair retained its disproved diagnosis")
				}
			}
			assertConnectionRecordFieldSet(t, before, fresh, allowed)
		})
	}
}

type repairedConnectionListSession struct{ calls int }

func (*repairedConnectionListSession) Close(context.Context) error { return nil }
func (s *repairedConnectionListSession) Call(_ context.Context, export string, payload []byte) ([]byte, error) {
	if export != "list" {
		return nil, pluginsystem.Failure("invalid_request", "")
	}
	s.calls++
	var input pluginSystemListInput
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, err
	}
	if input.Auth["apiKey"] != "stored-input" {
		return nil, pluginsystem.Failure("invalid_grant", pluginsystem.OperationList)
	}
	return json.Marshal(pluginSystemListOutput{Items: []pluginsystem.TrailSummary{}, HasMore: false})
}

func TestCronMissingSelectionThenHostRepairCanCompleteSync(t *testing.T) {
	f := newConnectionHostFixture(t)
	before := f.create(t, map[string]any{"apiKey": "stored-input"})
	before.Set("auth_context", "")
	before.Set("state", map[string]any{"oldCursor": "preserve-until-sync"})
	if err := f.app.Save(before); err != nil {
		t.Fatal(err)
	}
	actors := core.NewBaseCollection("activitypub_actors")
	actors.Fields.Add(&core.TextField{Name: "user"})
	if err := f.app.Save(actors); err != nil {
		t.Fatal(err)
	}
	actor := core.NewRecord(actors)
	actor.Set("user", f.owner.Id)
	if err := f.app.Save(actor); err != nil {
		t.Fatal(err)
	}
	plugin := pluginsystem.LocalPlugin{Manifest: f.manifest}
	// The actual Cron opening path writes the diagnosis before runtime/provider
	// work, exactly as it does for a legacy record with no selected context.
	if _, err := syncPluginInstance(context.Background(), f.app, nil, plugin, before); pluginsystem.PublicPluginError(err).Code != "invalid_configuration" {
		t.Fatalf("Cron missing selection produced %v", err)
	}
	failed, _ := f.app.FindRecordById("plugin_instances", before.Id)
	diagnosis := pluginsystem.JSONMapFromRecord(failed, "last_error")
	if diagnosis["code"] != "invalid_configuration" || diagnosis["authContext"] != "" || diagnosis["connectionGeneration"] != before.GetString("connection_generation") {
		t.Fatal("Cron fixture did not produce the actual current empty-context diagnosis")
	}
	status, _, body := f.patch(t, failed, map[string]any{"authContext": "api", "expected_connection_generation": failed.GetString("connection_generation")})
	if status != http.StatusOK {
		t.Fatalf("host selection repair %d: %s", status, body)
	}
	repaired, _ := f.app.FindRecordById("plugin_instances", before.Id)
	assertConnectionRecordFieldSet(t, failed, repaired, map[string]bool{"auth_context": true, "status": true, "last_error": true})
	if repaired.GetString("status") != "configured" || len(pluginsystem.JSONMapFromRecord(repaired, "last_error")) != 0 {
		t.Fatal("host repair retained the obsolete empty-context error")
	}
	auth, err := decryptedActiveInstanceAuth(plugin, repaired)
	if err != nil {
		t.Fatal(err)
	}
	// Only provider output is stubbed. Capability C+ checks, input projection and
	// final status/state CAS use the production sync writers on the repaired DB.
	session := &repairedConnectionListSession{}
	if _, err := syncPluginCapability(context.Background(), f.app, nil, &pluginSyncRuntimeSession{session: session}, plugin, pluginsystem.CapabilityManifest{Name: "list_routes", Export: "list"}, pluginsystem.CapabilityManifest{Name: "get_route_detail", Export: "detail"}, repaired, actor, auth, nil, nil, false, false); err != nil {
		t.Fatal(err)
	}
	if err := completePluginSync(f.app, repaired, time.Now()); err != nil {
		t.Fatal(err)
	}
	fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
	if session.calls != 1 || fresh.GetString("status") != "configured" || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || fresh.GetDateTime("last_sync_at").IsZero() || fresh.GetString("connection_generation") != before.GetString("connection_generation") || fresh.GetString("oauth_token_revision") != before.GetString("oauth_token_revision") {
		t.Fatal("repaired connection could not complete a successful sync")
	}
}

func TestUnattributedAPIAccessTokenIsConsumedWhenSelectingOAuth(t *testing.T) {
	for _, sourceType := range []string{pluginsystem.AuthTypeBearer, pluginsystem.AuthTypeAPIKey} {
		t.Run(sourceType, func(t *testing.T) {
			f := newConnectionHostFixture(t)
			f.manifest.Auth.Contexts = map[string]pluginsystem.AuthContext{
				"input": {Type: sourceType, SecretField: "accessToken"},
				"oauth": {Type: pluginsystem.AuthTypeOAuth2, Fields: []string{"clientId", "clientSecret"}, AuthorizationURL: "https://example.test/auth", TokenURL: "https://example.test/token", Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModeHost}},
			}
			f.installed.Set("manifest", f.manifest)
			if err := f.app.Save(f.installed); err != nil {
				t.Fatal(err)
			}
			status, result, body := f.request(t, http.MethodPost, "/plugins/instances", map[string]any{"plugin_id": f.manifest.ID, "authContext": "input", "auth": map[string]any{"accessToken": "api-input"}, "enabled": true})
			if status != http.StatusOK {
				t.Fatalf("API create %d: %s", status, body)
			}
			before, _ := f.app.FindRecordById("plugin_instances", result["id"].(string))
			before.Set("state", map[string]any{"cursor": "old-api-account"})
			before.Set("last_error", map[string]any{"code": "historical"})
			before.Set("retry_not_before", "2030-01-01 00:00:00.000Z")
			if err := f.app.Save(before); err != nil {
				t.Fatal(err)
			}
			status, _, body = f.patch(t, before, map[string]any{"authContext": "oauth", "auth": map[string]any{"clientId": "new-client", "clientSecret": "new-client-secret"}, "expected_connection_generation": before.GetString("connection_generation")})
			if status != http.StatusOK {
				t.Fatalf("OAuth switch %d: %s", status, body)
			}
			fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
			auth := pluginsystem.JSONMapFromRecord(fresh, "auth")
			for _, field := range append(append([]string{"accessToken", "refreshToken"}, pluginsystem.InternalOAuthMetadataFields()...), pluginsystem.InternalOAuthTransientFields()...) {
				if _, present := auth[field]; present {
					t.Errorf("OAuth switch retained API/grant field %s", field)
				}
			}
			secret, err := security.Decrypt(auth["clientSecret"].(string), oauthMetadataTestKey)
			if err != nil || string(secret) != "new-client-secret" || auth["clientId"] != "new-client" {
				t.Fatal("OAuth switch lost explicit new client credentials")
			}
			if fresh.GetBool("enabled") || fresh.GetString("status") != "disabled" || fresh.GetString("connection_generation") == before.GetString("connection_generation") || fresh.GetString("oauth_token_revision") == before.GetString("oauth_token_revision") {
				t.Fatal("OAuth switch retained API connection/revision or remained enabled without a grant")
			}
			assertConnectionRecordFieldSet(t, before, fresh, map[string]bool{"auth_context": true, "auth": true, "enabled": true, "status": true, "connection_generation": true, "oauth_token_revision": true, "state": true, "last_error": true, "retry_not_before": true})
		})
	}
}

func TestNonOAuthContextSwitchRequiresItsOwnCredentialsToRemainEnabled(t *testing.T) {
	for _, targetType := range []string{pluginsystem.AuthTypeBearer, pluginsystem.AuthTypeAPIKey, pluginsystem.AuthTypeSession} {
		for _, provide := range []bool{false, true} {
			t.Run(targetType+map[bool]string{false: "-missing", true: "-explicit"}[provide], func(t *testing.T) {
				f := newConnectionHostFixture(t)
				before := f.create(t, map[string]any{"apiKey": "old-api-input"})
				context := pluginsystem.AuthContext{Type: targetType, SecretField: "newToken"}
				newCredentials := map[string]any{"newToken": "new-input"}
				if targetType == pluginsystem.AuthTypeSession {
					context = pluginsystem.AuthContext{Type: targetType, Fields: []string{"email", "password"}, SecretFields: []string{"password"}, Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModePlugin, Function: "login"}}
					newCredentials = map[string]any{"email": "new@example.test", "password": "new-password"}
				}
				f.manifest.Auth.Contexts["target"] = context
				f.installed.Set("manifest", f.manifest)
				if err := f.app.Save(f.installed); err != nil {
					t.Fatal(err)
				}
				patch := map[string]any{"authContext": "target", "expected_connection_generation": before.GetString("connection_generation")}
				if provide {
					patch["auth"] = newCredentials
				} else {
					patch["enabled"] = true
					status, _, body := f.patch(t, before, patch)
					if status != http.StatusBadRequest {
						t.Fatalf("unready switch + activate %d: %s", status, body)
					}
					fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
					assertConnectionRecordFieldSet(t, before, fresh, nil)
					delete(patch, "enabled")
				}
				status, _, body := f.patch(t, before, patch)
				if status != http.StatusOK {
					t.Fatalf("context switch %d: %s", status, body)
				}
				fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
				if fresh.GetBool("enabled") != provide || fresh.GetString("status") != map[bool]string{false: "disabled", true: "configured"}[provide] || fresh.GetString("connection_generation") == before.GetString("connection_generation") {
					t.Fatal("non-OAuth switch did not apply credential readiness")
				}
				if pluginsystem.JSONMapFromRecord(fresh, "auth")["apiKey"] != pluginsystem.JSONMapFromRecord(before, "auth")["apiKey"] {
					t.Fatal("context switch modified separate inactive credentials")
				}
				assertConnectionRecordFieldSet(t, before, fresh, map[string]bool{"auth_context": true, "auth": true, "enabled": true, "status": true, "connection_generation": true, "state": true, "last_error": true, "retry_not_before": true})
			})
		}
	}
}

func TestRemovedAPIContextCannotBecomeSoleOAuthGrant(t *testing.T) {
	for _, sourceType := range []string{pluginsystem.AuthTypeAPIKey, pluginsystem.AuthTypeBearer} {
		for _, emptyMarker := range []bool{false, true} {
			t.Run(sourceType+map[bool]string{false: "-no-provenance", true: "-empty-provenance"}[emptyMarker], func(t *testing.T) {
				f := newConnectionHostFixture(t)
				f.manifest.Auth.Contexts = map[string]pluginsystem.AuthContext{"input": {Type: sourceType, SecretField: "accessToken"}}
				f.installed.Set("manifest", f.manifest)
				if err := f.app.Save(f.installed); err != nil {
					t.Fatal(err)
				}
				status, result, body := f.request(t, http.MethodPost, "/plugins/instances", map[string]any{"plugin_id": f.manifest.ID, "authContext": "input", "auth": map[string]any{"accessToken": "actual-api-key"}, "enabled": true})
				if status != http.StatusOK {
					t.Fatalf("API create %d: %s", status, body)
				}
				before, _ := f.app.FindRecordById("plugin_instances", result["id"].(string))
				if emptyMarker {
					auth := pluginsystem.JSONMapFromRecord(before, "auth")
					auth["oauthContext"] = ""
					before.Set("auth", auth)
					if err := f.app.Save(before); err != nil {
						t.Fatal(err)
					}
				}
				// Replace the entire manifest's auth choices. Looking up the old
				// context now yields no type; the new sole OAuth mode must not
				// turn the old API credential into an unattributed legacy grant.
				f.manifest.Auth.Contexts = map[string]pluginsystem.AuthContext{"oauth": {Type: pluginsystem.AuthTypeOAuth2, Fields: []string{"clientId", "clientSecret"}, AuthorizationURL: "https://example.test/auth", TokenURL: "https://example.test/token", Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModeHost}}}
				f.installed.Set("manifest", f.manifest)
				if err := f.app.Save(f.installed); err != nil {
					t.Fatal(err)
				}
				patch := map[string]any{"authContext": "oauth", "expected_connection_generation": before.GetString("connection_generation"), "enabled": true}
				status, _, body = f.patch(t, before, patch)
				if status != http.StatusBadRequest {
					t.Fatalf("API key activated as OAuth grant %d: %s", status, body)
				}
				fresh, _ := f.app.FindRecordById("plugin_instances", before.Id)
				assertConnectionRecordFieldSet(t, before, fresh, nil)
				delete(patch, "enabled")
				status, _, body = f.patch(t, before, patch)
				if status != http.StatusOK {
					t.Fatalf("OAuth switch %d: %s", status, body)
				}
				fresh, _ = f.app.FindRecordById("plugin_instances", before.Id)
				auth := pluginsystem.JSONMapFromRecord(fresh, "auth")
				if _, exists := auth["accessToken"]; exists || fresh.GetBool("enabled") || fresh.GetString("status") != "disabled" || fresh.GetString("connection_generation") == before.GetString("connection_generation") || fresh.GetString("oauth_token_revision") == before.GetString("oauth_token_revision") {
					t.Fatal("removed API context's key became a sole-context OAuth grant")
				}
				assertConnectionRecordFieldSet(t, before, fresh, map[string]bool{"auth_context": true, "auth": true, "enabled": true, "status": true, "connection_generation": true, "oauth_token_revision": true, "state": true, "last_error": true, "retry_not_before": true})
			})
		}
	}
}
