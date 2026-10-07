package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const pluginHostRoutePrivateError = "token=route-token-secret path=/private/route-credential-secret/file.gpx meili=https://meili-private.invalid/?apiKey=route-meili-secret"

// Inject only a failing database boundary; requests, authentication, database
// driver errors, remap transactions and the original BatchHandler stay real.
type pluginHostRouteFailureApp struct {
	core.App
	filterCollection string
	filterFailureAt  int
	filterCalls      int
	recordCollection string
}

func (a *pluginHostRouteFailureApp) databaseFailure() error {
	_, err := a.App.DB().NewQuery("SELECT route_db_secret_sentinel()").Execute()
	return err
}

func (a *pluginHostRouteFailureApp) FindRecordsByFilter(collection any, filter, sort string, limit, offset int, params ...dbx.Params) ([]*core.Record, error) {
	if collection == a.filterCollection {
		a.filterCalls++
		if a.filterCalls == a.filterFailureAt {
			return nil, a.databaseFailure()
		}
	}
	return a.App.FindRecordsByFilter(collection, filter, sort, limit, offset, params...)
}

func (a *pluginHostRouteFailureApp) FindRecordById(collection any, id string, filters ...func(*dbx.SelectQuery) error) (*core.Record, error) {
	if collection == a.recordCollection {
		return nil, a.databaseFailure()
	}
	return a.App.FindRecordById(collection, id, filters...)
}

func (a *pluginHostRouteFailureApp) RunInTransaction(fn func(core.App) error) error {
	return a.App.RunInTransaction(func(tx core.App) error {
		copy := *a
		copy.App = tx
		return fn(&copy)
	})
}

func pluginHostRouteRequest(t *testing.T, app core.App, owner *core.Record, method, path, body string, handlerFunc func(*core.RequestEvent) error) (int, map[string]any) {
	t.Helper()
	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	router.Route(method, path, handlerFunc)
	handler, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if owner != nil {
		token, err := owner.NewAuthToken()
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response %d: %s: %v", response.Code, response.Body.String(), err)
	}
	return response.Code, result
}

func assertPluginHostRouteLogs(t *testing.T, app core.App, method, path string, status int, body map[string]any, action, cause string, private ...string) {
	t.Helper()
	if body["status"] != float64(status) {
		t.Errorf("API status: got %#v, want %d", body, status)
	}
	data, ok := body["data"].(map[string]any)
	if !ok || len(data) != 0 {
		t.Errorf("host error must have empty API data, without a plugin/provider code: %#v", body)
	}
	logs := waitForPluginRequestLogsForMethod(t, app, method, path, 1)
	diagnostics := 0
	for _, entry := range logs {
		if entry.Message == "plugin operation failed" {
			diagnostics++
			if entry.Data["action"] != action || entry.Data["operation"] != action || entry.Data["cause"] != cause {
				t.Errorf("unexpected diagnostic: %#v", entry)
			}
		}
	}
	wantDiagnostics := 1
	if cause == "" {
		wantDiagnostics = 0
	}
	if diagnostics != wantDiagnostics {
		t.Errorf("got %d safe diagnostics, want %d: %#v", diagnostics, wantDiagnostics, logs)
	}
	private = append(private, "route-token-secret", "/private/route-credential-secret", "route-meili-secret", "meili-private.invalid", "route_db_secret_sentinel")
	for _, value := range []any{body, logs} {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		for _, secret := range private {
			if strings.Contains(string(encoded), secret) {
				t.Errorf("API/persisted PocketBase log leaked %q: %s", secret, encoded)
			}
		}
	}
}

func TestPluginsListHostFailuresDoNotLeakIntoPocketBaseRequestLogs(t *testing.T) {
	for _, testCase := range []string{"filesystem", "cache database", "list database", "cache deletion hook"} {
		t.Run(testCase, func(t *testing.T) {
			f := newInteractivePluginFixture(t, nil)
			var app core.App = f.app
			dir := t.TempDir()
			cause := "plugin_cache_sync"
			switch testCase {
			case "filesystem":
				dir = filepath.Join(dir, "private-plugin-path-secret")
				if err := os.WriteFile(dir, []byte("not a directory"), 0600); err != nil {
					t.Fatal(err)
				}
			case "cache database":
				app = &pluginHostRouteFailureApp{App: f.app, filterCollection: "installed_plugins", filterFailureAt: 1}
			case "list database":
				app = &pluginHostRouteFailureApp{App: f.app, filterCollection: "installed_plugins", filterFailureAt: 2}
				cause = "plugin_list"
			case "cache deletion hook":
				f.app.OnRecordDelete("installed_plugins").BindFunc(func(e *core.RecordEvent) error {
					return errors.New(pluginHostRoutePrivateError)
				})
			}
			enablePluginRequestLogs(t, f.app)
			const path = "/api/test/plugins"
			status, body := pluginHostRouteRequest(t, app, f.owner, http.MethodGet, path, "", func(e *core.RequestEvent) error {
				return pluginSystemPluginsList(e, dir)
			})
			if status != 400 || body["message"] != "Failed to load plugins." {
				t.Fatalf("status=%d body=%#v", status, body)
			}
			assertPluginHostRouteLogs(t, f.app, http.MethodGet, path, status, body, "plugins_list", cause, "private-plugin-path-secret", dir)
		})
	}
}

func newPluginCategoryRemapFixture(t *testing.T) (*interactivePluginFixture, *core.Record, *core.Record) {
	t.Helper()
	f := newInteractivePluginFixture(t, nil)
	categories := core.NewBaseCollection("categories")
	categories.Fields.Add(&core.TextField{Name: "name"})
	if err := f.app.Save(categories); err != nil {
		t.Fatal(err)
	}
	category := core.NewRecord(categories)
	category.Set("name", "Hiking")
	if err := f.app.Save(category); err != nil {
		t.Fatal(err)
	}
	trails := core.NewBaseCollection("trails")
	trails.Fields.Add(&core.RelationField{Name: "category", CollectionId: categories.Id, MaxSelect: 1}, &core.TextField{Name: "subcategory"})
	if err := f.app.Save(trails); err != nil {
		t.Fatal(err)
	}
	trail := core.NewRecord(trails)
	if err := f.app.Save(trail); err != nil {
		t.Fatal(err)
	}
	refs := core.NewBaseCollection("trail_external_reference")
	refs.Fields.Add(&core.TextField{Name: "user"}, &core.TextField{Name: "plugin_id"}, &core.TextField{Name: "trail"}, &core.TextField{Name: "external_id"}, &core.TextField{Name: "provider_category"})
	if err := f.app.Save(refs); err != nil {
		t.Fatal(err)
	}
	ref := core.NewRecord(refs)
	ref.Set("user", f.owner.Id)
	ref.Set("plugin_id", f.manifest.ID)
	ref.Set("trail", trail.Id)
	ref.Set("external_id", "provider-external-secret")
	ref.Set("provider_category", "Ride")
	if err := f.app.Save(ref); err != nil {
		t.Fatal(err)
	}
	return f, trail, category
}

func TestCategoryRemapHostFailuresDoNotLeakIntoPocketBaseRequestLogs(t *testing.T) {
	for _, action := range []string{"preview", "apply"} {
		for _, failure := range []string{"decoder", "instance database", "references database", "transaction database", "save hook", "after commit hook"} {
			if action == "preview" && (failure == "transaction database" || failure == "save hook" || failure == "after commit hook") {
				continue
			}
			t.Run(action+"/"+failure, func(t *testing.T) {
				f, trail, category := newPluginCategoryRemapFixture(t)
				var app core.App = f.app
				payload, err := json.Marshal(map[string]any{"instanceId": f.instance.Id, "config": map[string]any{"host": map[string]any{"categoryMapping": map[string]any{"Ride": category.Id}}}})
				if err != nil {
					t.Fatal(err)
				}
				bodyText := string(payload)
				cause := "database_lookup"
				wantStatus := 400
				hookCalls := 0
				switch failure {
				case "decoder":
					bodyText = `{"instanceId":"request-instance-secret","config":{"secret":"body-secret",`
					cause = "request_body_invalid"
				case "instance database":
					app = &pluginHostRouteFailureApp{App: f.app, recordCollection: "plugin_instances"}
					wantStatus = 404
				case "references database":
					app = &pluginHostRouteFailureApp{App: f.app, filterCollection: "trail_external_reference", filterFailureAt: 1}
				case "transaction database":
					app = &pluginHostRouteFailureApp{App: f.app, recordCollection: "trails"}
				case "save hook":
					cause = "database_save"
					f.app.OnRecordUpdate("trails").BindFunc(func(e *core.RecordEvent) error {
						hookCalls++
						return errors.New(pluginHostRoutePrivateError)
					})
				case "after commit hook":
					cause = "database_save"
					f.app.OnRecordAfterUpdateSuccess("trails").BindFunc(func(e *core.RecordEvent) error {
						hookCalls++
						return errors.New(pluginHostRoutePrivateError)
					})
				}
				enablePluginRequestLogs(t, f.app)
				path := "/api/test/remap/" + action
				handler := PluginSystemCategoryRemapPreview
				if action == "apply" {
					handler = PluginSystemCategoryRemapApply
				}
				status, body := pluginHostRouteRequest(t, app, f.owner, http.MethodPost, path, bodyText, handler)
				if status != wantStatus {
					t.Fatalf("status=%d want=%d body=%#v", status, wantStatus, body)
				}
				assertPluginHostRouteLogs(t, f.app, http.MethodPost, path, status, body, "category_remap_"+action, cause, "request-instance-secret", "body-secret", "provider-external-secret")
				if failure == "save hook" || failure == "after commit hook" {
					if hookCalls != 1 {
						t.Fatalf("did not reach actual save hook: calls=%d", hookCalls)
					}
					stored, err := f.app.FindRecordById("trails", trail.Id)
					if err != nil {
						t.Fatal(err)
					}
					// This patch only sanitizes failures. PocketBase's existing
					// commit/after-hook semantics are intentionally unchanged.
					wantCategory := ""
					if failure == "after commit hook" {
						wantCategory = category.Id
					}
					if stored.GetString("category") != wantCategory {
						t.Errorf("stored category=%q, want %q", stored.GetString("category"), wantCategory)
					}
				}
			})
		}
	}
}

func TestCategoryRemapLocalStatusErrorsStayPlainAndPrivate(t *testing.T) {
	for _, action := range []string{"preview", "apply"} {
		for _, failure := range []string{"unauthenticated", "missing instanceId", "absent instance", "foreign owner"} {
			t.Run(action+"/"+failure, func(t *testing.T) {
				f := newInteractivePluginFixture(t, nil)
				owner := f.owner
				wantStatus := 404
				bodyText := `{"instanceId":"request-instance-secret"}`
				cause := "database_lookup"
				switch failure {
				case "unauthenticated":
					owner = nil
					wantStatus = 401
					cause = ""
				case "missing instanceId":
					bodyText = `{}`
					wantStatus = 400
					cause = ""
				case "foreign owner":
					f.instance.Set("user", "")
					if err := f.app.Save(f.instance); err != nil {
						t.Fatal(err)
					}
					bodyText = `{"instanceId":"` + f.instance.Id + `"}`
					cause = ""
				}
				enablePluginRequestLogs(t, f.app)
				path := "/api/test/remap/" + action
				handler := PluginSystemCategoryRemapPreview
				if action == "apply" {
					handler = PluginSystemCategoryRemapApply
				}
				status, body := pluginHostRouteRequest(t, f.app, owner, http.MethodPost, path, bodyText, handler)
				if status != wantStatus {
					t.Fatalf("status=%d want=%d body=%#v", status, wantStatus, body)
				}
				assertPluginHostRouteLogs(t, f.app, http.MethodPost, path, status, body, "category_remap_"+action, cause, "request-instance-secret", f.instance.Id)
			})
		}
	}
}
