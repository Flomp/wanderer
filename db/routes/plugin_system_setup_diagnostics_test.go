package routes

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
	"pocketbase/pluginsystem"
)

const setupDiagnosticPath = "/api/test/plugins"
const setupURLSecret = "manifest-url-secret-sentinel"
const setupPathSecret = "plugin-parent-path-secret-sentinel"

func TestPluginSetupDiagnosticsHealthyIdentityRenameChainKeepsBothRecords(t *testing.T) {
	for _, newIdentityFirst := range []bool{false, true} {
		t.Run(map[bool]string{false: "existing identity discovered first", true: "new identity discovered first"}[newIdentityFirst], func(t *testing.T) {
			f := newSetupDiagnosticFixture(t)
			f.manifest.ID = "identity-first"
			firstPath := f.bundle(t, "a-directory", "")
			f.manifest.ID = "identity-second"
			secondPath := f.bundle(t, "b-directory", "")
			f.request(t, http.StatusOK)
			retainedID, retainedPath, newPath := "identity-second", firstPath, secondPath
			if newIdentityFirst {
				retainedID, retainedPath, newPath = "identity-first", secondPath, firstPath
			}
			retained, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", retainedID)
			if err != nil {
				t.Fatal(err)
			}
			retained.Set("config", map[string]any{"plugin": map[string]any{"admin_value": "retained"}})
			if err := f.app.Save(retained); err != nil {
				t.Fatal(err)
			}
			f.manifest.ID = retainedID
			f.bundle(t, filepath.Base(retainedPath), "")
			f.manifest.ID = "identity-third"
			f.bundle(t, filepath.Base(newPath), "")
			f.request(t, http.StatusOK)
			if count, err := f.app.CountRecords("installed_plugins"); err != nil || count != 2 {
				t.Fatalf("discovery lost an identity in rename chain: %d/%v", count, err)
			}
			fresh, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", retainedID)
			if err != nil || fresh.Id != retained.Id || fresh.GetString("path") != retainedPath {
				t.Fatalf("retained identity claimed by new plugin: %v/%v", fresh, err)
			}
			config := pluginsystem.JSONMapFromRecord(fresh, "config")
			if plugin, ok := config["plugin"].(map[string]any); !ok || plugin["admin_value"] != "retained" {
				t.Fatalf("retained identity lost its administrator config: %#v", config)
			}
			third, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", "identity-third")
			if err != nil || third.Id == retained.Id || third.GetString("path") != newPath || third.GetString("status") != "available" {
				t.Fatalf("new identity shares or loses the retained row: %v/%v", third, err)
			}
			f.request(t, http.StatusOK)
			if count, err := f.app.CountRecords("installed_plugins"); err != nil || count != 2 {
				t.Fatalf("repeat discovery changed chain result: %d/%v", count, err)
			}
			f.warnings(t, 0)
		})
	}
}

func TestPluginSetupDiagnosticsHealthyIdentitySwapKeepsBothCacheRecords(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	f.manifest.ID = "identity-first"
	firstPath := f.bundle(t, "first-directory", "")
	f.manifest.ID = "identity-second"
	secondPath := f.bundle(t, "second-directory", "")
	f.request(t, http.StatusOK)
	before := map[string]string{}
	for _, id := range []string{"identity-first", "identity-second"} {
		row, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", id)
		if err != nil {
			t.Fatal(err)
		}
		before[id] = row.Id
		row.Set("config", map[string]any{"plugin": map[string]any{"admin_value": id}})
		if err := f.app.Save(row); err != nil {
			t.Fatal(err)
		}
	}
	f.manifest.ID = "identity-second"
	f.bundle(t, "first-directory", "")
	f.manifest.ID = "identity-first"
	f.bundle(t, "second-directory", "")
	f.request(t, http.StatusOK)
	for id, path := range map[string]string{"identity-first": secondPath, "identity-second": firstPath} {
		row, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", id)
		if err != nil {
			t.Fatalf("healthy identity disappeared during relocation: %s/%v", id, err)
		}
		if row.Id != before[id] || row.GetString("path") != path || row.GetString("status") != "available" {
			t.Fatalf("identity relocation lost record/path: %#v", row.FieldsData())
		}
		config := pluginsystem.JSONMapFromRecord(row, "config")
		if plugin, ok := config["plugin"].(map[string]any); !ok || plugin["admin_value"] != id {
			t.Fatalf("identity config moved to wrong plugin: %#v", config)
		}
	}
	f.request(t, http.StatusOK)
	if count, err := f.app.CountRecords("installed_plugins"); err != nil || count != 2 {
		t.Fatalf("repeat discovery changed cache cardinality: %d/%v", count, err)
	}
	f.warnings(t, 0)
}

func TestPluginSetupDiagnosticsBrokenFormerIdentityCannotHideHealthyBundle(t *testing.T) {
	for _, discoverFailureFirst := range []bool{false, true} {
		t.Run(map[bool]string{false: "first broken discovery", true: "previously cached failure"}[discoverFailureFirst], func(t *testing.T) {
			f := newSetupDiagnosticFixture(t)
			oldPath := f.bundle(t, "former-directory", "")
			f.request(t, http.StatusOK)
			old, err := f.app.FindFirstRecordByData("installed_plugins", "path", oldPath)
			if err != nil {
				t.Fatal(err)
			}
			old.Set("config", map[string]any{"plugin": map[string]any{"admin_value": "keep"}})
			if err := f.app.Save(old); err != nil {
				t.Fatal(err)
			}
			f.bundle(t, "former-directory", "https://provider.example/%zz?token="+setupURLSecret)
			if discoverFailureFirst {
				f.request(t, http.StatusOK)
			}
			healthyPath := f.bundle(t, "healthy-directory", "")
			f.request(t, http.StatusOK)
			f.request(t, http.StatusOK)
			rows, err := f.app.FindAllRecords("installed_plugins")
			if err != nil || len(rows) != 2 {
				t.Fatalf("discovery cache rows=%v, error=%v", rows, err)
			}
			healthy, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", f.manifest.ID)
			if err != nil {
				t.Fatal(err)
			}
			if healthy.GetString("path") != healthyPath || healthy.GetString("status") != "available" {
				t.Fatalf("healthy identity claimed by broken directory: %#v", healthy.FieldsData())
			}
			config := pluginsystem.JSONMapFromRecord(healthy, "config")
			if plugin, ok := config["plugin"].(map[string]any); !ok || plugin["admin_value"] != "keep" {
				t.Fatalf("administrator config lost: %#v", config)
			}
			broken, err := f.app.FindFirstRecordByData("installed_plugins", "path", oldPath)
			if err != nil || broken.GetString("status") != "error" || broken.GetString("plugin_id") == f.manifest.ID {
				t.Fatalf("broken directory was not kept separate: %v/%v", broken, err)
			}
			warnings := f.warnings(t, 1)
			if !discoverFailureFirst && len(warnings) == 1 && warnings[0].Data["plugin"] == f.manifest.ID {
				t.Fatal("broken directory warning blamed the healthy identity")
			}
			response := f.get()
			var public struct {
				Items []pluginsystem.PluginInfo `json:"items"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &public); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, item := range public.Items {
				if item.ID == f.manifest.ID && item.Status == "available" {
					found = true
				}
			}
			if !found {
				t.Fatalf("healthy plugin hidden from API: %s", response.Body.String())
			}
		})
	}
}

func TestPluginSetupDiagnosticsRestoreDisplacedValidatedIdentity(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	originalPath := f.bundle(t, "former-directory", "")
	f.request(t, http.StatusOK)
	original, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", f.manifest.ID)
	if err != nil {
		t.Fatal(err)
	}
	original.Set("config", map[string]any{"plugin": map[string]any{"admin_value": "keep"}})
	if err := f.app.Save(original); err != nil {
		t.Fatal(err)
	}
	f.bundle(t, "former-directory", "https://provider.example/%zz?token="+setupURLSecret)
	competitorPath := f.bundle(t, "healthy-directory", "")
	f.request(t, http.StatusOK)
	displaced, err := f.app.FindRecordById("installed_plugins", original.Id)
	if err != nil || displaced.GetString("plugin_id") == f.manifest.ID || displaced.GetString("status") != "error" {
		t.Fatalf("original diagnostic was not displaced: %v/%v", displaced, err)
	}
	if err := os.RemoveAll(competitorPath); err != nil {
		t.Fatal(err)
	}
	f.request(t, http.StatusOK)
	restored, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", f.manifest.ID)
	if err != nil || restored.Id != original.Id || restored.GetString("path") != originalPath || restored.GetString("status") != "error" {
		t.Fatalf("validated diagnostic identity not restored: %v/%v", restored, err)
	}
	if !strings.Contains(restored.GetString("config"), "keep") {
		t.Fatal("identity restoration lost administrator configuration")
	}
	if _, err := pluginsystem.LoadInstalledPlugin(f.app, f.dir, f.manifest.ID); err == nil {
		t.Fatal("restored diagnostic made broken bundle executable")
	}
	var saves atomic.Int32
	f.app.OnRecordUpdate("installed_plugins").BindFunc(func(e *core.RecordEvent) error { saves.Add(1); return e.Next() })
	f.request(t, http.StatusOK)
	if saves.Load() != 0 {
		t.Fatal("restored identity did not remain stable")
	}
	f.bundle(t, "former-directory", "")
	f.request(t, http.StatusOK)
	healthy, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", f.manifest.ID)
	if err != nil || healthy.Id != original.Id || healthy.GetString("status") != "available" {
		t.Fatalf("bundle repair did not reuse restored identity: %v/%v", healthy, err)
	}
}

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

// Signal that the second manager finished its read plan and is about to
// acquire the writer. The insert hook separately detects an unsynchronized
// competitor, making omission of RunInTransaction fail deterministically.
type setupTransactionProbeApp struct {
	core.App
	transactions atomic.Int32
	second       chan struct{}
}

func (app *setupTransactionProbeApp) RunInTransaction(fn func(core.App) error) error {
	if app.transactions.Add(1) == 2 {
		close(app.second)
	}
	return app.App.RunInTransaction(fn)
}

func TestPluginSetupDiagnosticsParallelManagersLogOnce(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	f.bundle(t, "setup-log-plugin", "https://provider.example/%zz?token="+setupURLSecret)
	inserted, duplicate, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var creates atomic.Int32
	f.app.OnRecordCreate("installed_plugins").BindFunc(func(e *core.RecordEvent) error {
		if creates.Add(1) == 1 {
			close(inserted)
			<-release
		} else {
			close(duplicate)
		}
		return e.Next()
	})
	app := &setupTransactionProbeApp{App: f.app, second: make(chan struct{})}
	results := make(chan error, 2)
	go func() { results <- pluginsystem.NewManager(app, f.dir).SyncInstalledPlugins(context.Background()) }()
	select {
	case <-inserted:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("first cache insert did not reach barrier")
	}
	go func() { results <- pluginsystem.NewManager(app, f.dir).SyncInstalledPlugins(context.Background()) }()
	select {
	case <-app.second:
	case <-duplicate:
		t.Error("second manager inserted against a stale snapshot outside the serialized writer")
	case <-time.After(5 * time.Second):
		t.Error("second manager did not reach transaction acquisition")
	}
	close(release)
	for range 2 {
		if err := <-results; err != nil {
			t.Error(err)
		}
	}
	if got := creates.Load(); got != 1 {
		t.Errorf("cache insert hooks=%d, want one", got)
	}
	f.warnings(t, 1)
}

func TestPluginSetupDiagnosticsUnchangedCacheDoesNotAcquireWriter(t *testing.T) {
	for _, populated := range []bool{false, true} {
		t.Run(map[bool]string{false: "empty directory", true: "unchanged installed bundle"}[populated], func(t *testing.T) {
			f := newSetupDiagnosticFixture(t)
			if populated {
				f.bundle(t, "directory-alias", "")
			}
			f.request(t, http.StatusOK)
			records, err := f.app.FindAllRecords("installed_plugins")
			if err != nil {
				t.Fatal(err)
			}
			var before string
			if len(records) == 1 {
				before = records[0].GetString("updated")
				records[0].Set("config", map[string]any{"plugin": map[string]any{"custom": "preserved"}})
				if err := f.app.Save(records[0]); err != nil {
					t.Fatal(err)
				}
				// Reconcile any newly missing manifest defaults before measuring a
				// genuinely unchanged cache with an open writer.
				f.request(t, http.StatusOK)
				fresh, err := f.app.FindRecordById("installed_plugins", records[0].Id)
				if err != nil {
					t.Fatal(err)
				}
				before = fresh.GetString("updated")
			}
			var saves atomic.Int32
			f.app.OnRecordUpdate("installed_plugins").BindFunc(func(e *core.RecordEvent) error { saves.Add(1); return e.Next() })
			entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
			go func() {
				done <- f.app.RunInTransaction(func(core.App) error { close(entered); <-release; return nil })
			}()
			<-entered
			response := make(chan *httptest.ResponseRecorder, 1)
			go func() { response <- f.get() }()
			select {
			case result := <-response:
				f.requests++
				if result.Code != http.StatusOK {
					t.Errorf("GET=%d: %s", result.Code, result.Body.String())
				}
			case <-time.After(2 * time.Second):
				close(release)
				<-done
				<-response
				t.Fatal("unchanged GET waited behind a writer")
			}
			close(release)
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			if saves.Load() != 0 {
				t.Fatal("unchanged listing ran update hooks")
			}
			if len(records) == 1 {
				fresh, err := f.app.FindRecordById("installed_plugins", records[0].Id)
				if err != nil {
					t.Fatal(err)
				}
				if fresh.GetString("updated") != before || !strings.Contains(fresh.GetString("config"), "preserved") {
					t.Fatal("unchanged cache changed timestamp or user configuration")
				}
			}
		})
	}
}

func TestPluginSetupDiagnosticsSortInvalidAuthContexts(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	dir := f.bundle(t, "directory-alias", "")
	manifest := f.manifest
	manifest.Auth.Contexts = map[string]pluginsystem.AuthContext{
		"z-invalid": {Type: "invalid-z"}, "a-invalid": {Type: "invalid-a"},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	for range 40 {
		f.request(t, http.StatusOK)
	}
	f.warnings(t, 1)
}

func TestPluginSetupDiagnosticsStableIdentityCannotExecuteInvalidBundle(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	f.bundle(t, "directory-alias", "")
	f.request(t, http.StatusOK)
	f.bundle(t, "directory-alias", "https://provider.example/%zz?token="+setupURLSecret)
	f.request(t, http.StatusOK)
	f.bundle(t, "directory-alias", "https://provider.example/%zz?token="+setupURLSecret+"-changed")
	f.request(t, http.StatusOK)
	for _, warning := range f.warnings(t, 2) {
		if warning.Data["plugin"] != f.manifest.ID {
			t.Errorf("validated identity lost: %#v", warning.Data)
		}
	}
	if _, err := pluginsystem.LoadInstalledPlugin(f.app, f.dir, f.manifest.ID); err == nil {
		t.Fatal("invalid bundle remained executable through its preserved manifest")
	}
}

func TestPluginSetupDiagnosticsDirectoryCollisionDoesNotClaimHealthyPlugin(t *testing.T) {
	f := newSetupDiagnosticFixture(t)
	healthyDir := f.bundle(t, "healthy-directory", "")
	f.request(t, http.StatusOK)
	healthy, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", f.manifest.ID)
	if err != nil {
		t.Fatal(err)
	}
	healthy.Set("config", map[string]any{"plugin": map[string]any{"custom": "healthy-settings"}})
	if err := f.app.Save(healthy); err != nil {
		t.Fatal(err)
	}
	brokenDir := f.bundle(t, f.manifest.ID, "https://provider.example/%zz?token="+setupURLSecret)
	for range 3 {
		f.request(t, http.StatusOK)
	}
	warnings := f.warnings(t, 1)
	if len(warnings) == 1 && warnings[0].Data["plugin"] == f.manifest.ID {
		t.Fatal("broken directory warning blamed healthy plugin")
	}
	fresh, err := f.app.FindRecordById("installed_plugins", healthy.Id)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.GetString("path") != healthyDir || fresh.GetString("status") != "available" || !strings.Contains(fresh.GetString("config"), "healthy-settings") {
		t.Fatal("broken directory overwrote healthy cache")
	}
	broken, err := f.app.FindFirstRecordByData("installed_plugins", "path", brokenDir)
	if err != nil || broken.Id == healthy.Id || broken.GetString("plugin_id") == f.manifest.ID {
		t.Fatalf("broken directory was not cached independently: %v", err)
	}
	if _, err := pluginsystem.LoadInstalledPlugin(f.app, f.dir, f.manifest.ID); err != nil {
		t.Fatal("healthy bundle unavailable", err)
	}
	f.bundle(t, f.manifest.ID, "https://provider.example/%zz?token="+setupURLSecret+"-changed")
	f.request(t, http.StatusOK)
	warnings = f.warnings(t, 2)
	for _, warning := range warnings {
		if warning.Data["plugin"] != broken.GetString("plugin_id") {
			t.Fatalf("changed failure cause blamed healthy identity: %#v", warning.Data)
		}
	}
	if err := os.RemoveAll(brokenDir); err != nil {
		t.Fatal(err)
	}
	f.request(t, http.StatusOK)
	if _, err := f.app.FindRecordById("installed_plugins", broken.Id); err == nil {
		t.Fatal("removed broken cache row survived")
	}
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
