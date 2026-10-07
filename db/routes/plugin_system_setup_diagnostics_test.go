package routes

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
	"pocketbase/pluginsystem"
)

const setupDiagnosticPath = "/api/test/plugins"
const setupURLSecret = "manifest-url-secret-sentinel"
const setupPathSecret = "plugin-parent-path-secret-sentinel"

type setupDiagnosticFixture struct {
	app        *pbtests.TestApp
	dir, token string
	manifest   pluginsystem.Manifest
	handler    http.Handler
	requests   int
}

// Keep the real PocketBase logger, router, transaction and installed cache
// schema. A minimal cache without status/error would miss deduplication bugs.
func newSetupDiagnosticFixture(t *testing.T) *setupDiagnosticFixture {
	t.Helper()
	app, err := pbtests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	installed := core.NewBaseCollection("installed_plugins")
	installed.Fields.Add(
		&core.TextField{Name: "plugin_id", Required: true, Min: 1, Max: 128, Pattern: `^[a-z0-9][a-z0-9_-]*$`},
		&core.TextField{Name: "name", Required: true, Min: 1, Max: 256},
		&core.SelectField{Name: "type", Required: true, MaxSelect: 1, Values: []string{"trails"}},
		&core.TextField{Name: "version", Required: true, Min: 1, Max: 64},
		&core.TextField{Name: "runtime", Required: true, Min: 1, Max: 32},
		&core.TextField{Name: "path"},
		&core.JSONField{Name: "manifest", Required: true, MaxSize: 2000000},
		&core.JSONField{Name: "config", MaxSize: 2000000},
		&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"available", "disabled", "error"}},
		&core.TextField{Name: "error"},
	)
	installed.AddIndex("idx_setup_plugin_id", true, "plugin_id", "")
	if err := app.Save(installed); err != nil {
		t.Fatal(err)
	}
	users := core.NewAuthCollection("setup_diagnostic_users")
	if err := app.Save(users); err != nil {
		t.Fatal(err)
	}
	owner := core.NewRecord(users)
	owner.SetEmail("setup-diagnostic@example.com")
	owner.SetPassword("setup-diagnostic-password")
	if err := app.Save(owner); err != nil {
		t.Fatal(err)
	}
	token, err := owner.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "plugins", "strava", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	f := &setupDiagnosticFixture{app: app, dir: filepath.Join(t.TempDir(), setupPathSecret), token: token}
	if err := json.Unmarshal(data, &f.manifest); err != nil {
		t.Fatal(err)
	}
	f.manifest.ID = "setup-log-plugin"
	f.manifest.Name = "Setup log plugin"
	if err := pluginsystem.ValidateManifest(f.manifest); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(f.dir, 0700); err != nil {
		t.Fatal(err)
	}
	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	router.GET(setupDiagnosticPath, func(e *core.RequestEvent) error { return pluginSystemPluginsList(e, f.dir) })
	f.handler, err = router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	enablePluginRequestLogs(t, app)
	return f
}

func (f *setupDiagnosticFixture) bundle(t *testing.T, name string, badURL string) string {
	t.Helper()
	dir := filepath.Join(f.dir, name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	manifest := f.manifest
	manifest.Auth.Contexts = make(map[string]pluginsystem.AuthContext, len(f.manifest.Auth.Contexts))
	for key, value := range f.manifest.Auth.Contexts {
		manifest.Auth.Contexts[key] = value
	}
	if badURL != "" {
		context := manifest.Auth.Contexts["oauth_access_token"]
		context.AuthorizationURL = badURL
		manifest.Auth.Contexts["oauth_access_token"] = context
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifest.Runtime.Entrypoint), []byte("\x00asm\x01\x00\x00\x00"), 0600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func (f *setupDiagnosticFixture) get() *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, setupDiagnosticPath, nil)
	request.Header.Set("Authorization", f.token)
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, request)
	return response
}

func (f *setupDiagnosticFixture) request(t *testing.T, wantStatus int) {
	t.Helper()
	response := f.get()
	f.requests++
	if response.Code != wantStatus {
		t.Errorf("status=%d, want %d: %s", response.Code, wantStatus, response.Body.String())
	}
	// Manifest URLs occur only in the invalid input, not the cached minimal
	// manifest. Preserve existing ordinary UI names while redacting diagnostics.
	for _, secret := range []string{setupURLSecret, setupPathSecret} {
		if strings.Contains(response.Body.String(), secret) {
			t.Errorf("response leaked %q: %s", secret, response.Body.String())
		}
	}
}

func (f *setupDiagnosticFixture) warnings(t *testing.T, want int) []*core.Log {
	t.Helper()
	logs := waitForPluginRequestLogsForMethod(t, f.app, http.MethodGet, setupDiagnosticPath, f.requests)
	encoded, err := json.Marshal(logs)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{setupURLSecret, setupPathSecret, "folder-id-secret", "cached-id-secret", "after-save-secret", "rollback-save-secret"} {
		if strings.Contains(string(encoded), secret) {
			t.Errorf("persisted setup/request log leaked %q: %s", secret, encoded)
		}
	}
	var warnings []*core.Log
	for _, entry := range logs {
		if entry.Message != "plugin setup error" {
			continue
		}
		warnings = append(warnings, entry)
		if entry.Level != int(slog.LevelWarn) || len(entry.Data) != 2 {
			t.Errorf("setup diagnostic must contain only identity and code: %#v", entry)
		}
		if _, ok := entry.Data["plugin"].(string); !ok {
			t.Errorf("missing bounded plugin identity: %#v", entry.Data)
		}
		code, _ := entry.Data["setup_code"].(string)
		if code != "manifest_invalid" && code != "runtime_entrypoint_missing" {
			t.Errorf("unexpected setup code: %#v", entry.Data)
		}
	}
	if len(warnings) != want {
		t.Errorf("setup warnings=%d, want %d: %s", len(warnings), want, encoded)
	}
	return warnings
}

func TestPluginSetupDiagnosticsTrackCommittedChanges(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	name := "setup-log-plugin"
	badURL := "https://provider.example/%zz?token=" + setupURLSecret
	dir := f.bundle(t, name, badURL)
	f.request(t, http.StatusOK)
	f.warnings(t, 1)
	// Each route call creates a new manager; unchanged failures stay silent.
	f.request(t, http.StatusOK)
	f.warnings(t, 1)
	// A different private cause with the same public code is a fresh change.
	f.bundle(t, name, badURL+"-changed")
	f.request(t, http.StatusOK)
	f.warnings(t, 2)
	cached, err := f.app.FindFirstRecordByData("installed_plugins", "path", dir)
	if err != nil {
		t.Fatal(err)
	}
	if cached.GetString("status") != "error" || !strings.Contains(cached.GetString("error"), setupURLSecret) {
		t.Fatal("cache lost the existing administrator diagnostic contract")
	}
	f.bundle(t, name, "")
	f.request(t, http.StatusOK)
	f.warnings(t, 2)
	cached, err = f.app.FindRecordById("installed_plugins", cached.Id)
	if err != nil {
		t.Fatal(err)
	}
	if cached.GetString("status") != "available" || cached.GetString("error") != "" {
		t.Fatal("recovery did not clear cached failure")
	}
	if err := os.Remove(filepath.Join(dir, f.manifest.Runtime.Entrypoint)); err != nil {
		t.Fatal(err)
	}
	f.request(t, http.StatusOK)
	warnings := f.warnings(t, 3)
	if len(warnings) == 3 && warnings[2].Data["setup_code"] != "runtime_entrypoint_missing" {
		t.Errorf("new setup class missing: %#v", warnings[2])
	}
	f.request(t, http.StatusOK)
	f.warnings(t, 3)
	f.bundle(t, name, "")
	f.request(t, http.StatusOK)
	f.bundle(t, name, badURL)
	f.request(t, http.StatusOK)
	f.warnings(t, 4)
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	f.request(t, http.StatusOK)
	f.warnings(t, 4)
	records, err := f.app.FindRecordsByFilter("installed_plugins", "", "", -1, 0)
	if err != nil || len(records) != 0 {
		t.Fatalf("removed bundle cache persists: %#v/%v", records, err)
	}
	f.bundle(t, name, badURL)
	f.request(t, http.StatusOK)
	warnings = f.warnings(t, 5)
	for _, entry := range warnings {
		if entry.Data["plugin"] != name {
			t.Errorf("identity=%#v, want %q", entry.Data["plugin"], name)
		}
	}
}

func TestPluginSetupDiagnosticsUseValidatedBoundedIdentities(t *testing.T) {
	for _, tc := range []struct {
		name, directory string
		seed            bool
		want            string
	}{
		{"cached validated manifest", "directory-alias", false, "setup-log-plugin"},
		{"unsafe directory and cache identity", "bundle-token=folder-id-secret", true, ""},
		{"long directory identity", strings.Repeat("a", 180), false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newSetupDiagnosticFixture(t)
			dir := f.bundle(t, tc.directory, "")
			if tc.want != "" {
				f.request(t, http.StatusOK)
			}
			if tc.seed {
				collection, err := f.app.FindCollectionByNameOrId("installed_plugins")
				if err != nil {
					t.Fatal(err)
				}
				record := core.NewRecord(collection)
				record.Set("plugin_id", "cached-id-secret")
				record.Set("name", "Incomplete cache")
				record.Set("type", "trails")
				record.Set("version", "unknown")
				record.Set("runtime", "wasm")
				record.Set("path", dir)
				record.Set("status", "available")
				record.Set("manifest", map[string]any{"id": "cached-id-secret", "type": "trails", "name": "Incomplete cache"})
				if err := f.app.Save(record); err != nil {
					t.Fatal(err)
				}
			}
			f.bundle(t, tc.directory, "https://provider.example/%zz?token="+setupURLSecret)
			f.request(t, http.StatusOK)
			warnings := f.warnings(t, 1)
			if len(warnings) != 1 {
				return
			}
			identity, _ := warnings[0].Data["plugin"].(string)
			if tc.want != "" {
				if identity != tc.want {
					t.Errorf("identity=%q, want validated manifest ID %q", identity, tc.want)
				}
			} else if !regexp.MustCompile(`^plugin-setup-error-[0-9a-f]{8}$`).MatchString(identity) {
				t.Errorf("unsafe/unbounded identity was not replaced by a bounded label: %q", identity)
			}
		})
	}
}

func TestPluginSetupDiagnosticsParallelManagersLogOnce(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	f.bundle(t, "setup-log-plugin", "https://provider.example/%zz?token="+setupURLSecret)
	const count = 8
	responses := make(chan *httptest.ResponseRecorder, count)
	for range count {
		go func() { responses <- f.get() }()
	}
	for range count {
		response := <-responses
		if response.Code != http.StatusOK {
			t.Errorf("parallel GET status=%d: %s", response.Code, response.Body.String())
		}
	}
	f.requests = count
	f.warnings(t, 1)
}

func TestPluginSetupDiagnosticsDoNotLogRolledBackCache(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	for _, name := range []string{"first-plugin", "second-plugin"} {
		f.bundle(t, name, "https://provider.example/%zz?token="+setupURLSecret)
	}
	hookID := f.app.OnRecordCreate("installed_plugins").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("name") == "second-plugin" {
			return errors.New("rollback-save-secret")
		}
		return e.Next()
	})
	f.request(t, http.StatusBadRequest)
	f.warnings(t, 0)
	records, err := f.app.FindRecordsByFilter("installed_plugins", "", "", -1, 0)
	if err != nil || len(records) != 0 {
		t.Fatalf("failed transaction retained cache records: %#v/%v", records, err)
	}
	f.app.OnRecordCreate("installed_plugins").Unbind(hookID)
	f.request(t, http.StatusOK)
	f.warnings(t, 2)
}

func TestPluginSetupDiagnosticsLogActualCommitDespiteAfterSuccessError(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	f.bundle(t, "setup-log-plugin", "https://provider.example/%zz?token="+setupURLSecret)
	f.app.OnRecordAfterCreateSuccess("installed_plugins").BindFunc(func(e *core.RecordEvent) error { return errors.New("after-save-secret") })
	f.request(t, http.StatusBadRequest)
	f.warnings(t, 1)
	record, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", "setup-log-plugin")
	if err != nil || record.GetString("status") != "error" {
		t.Fatalf("actual SQL commit missing: %#v/%v", record, err)
	}
	f.request(t, http.StatusOK)
	f.warnings(t, 1)
}
