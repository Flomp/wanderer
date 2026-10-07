package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"pocketbase/pluginsystem"
)

func pluginInputRequestLogHandler(t *testing.T, app core.App, path string, handle func(*core.RequestEvent) error) http.Handler {
	t.Helper()
	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	router.POST(path, handle)
	handler, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func pluginInputRequestLogPost(t *testing.T, handler http.Handler, token, path string, body io.Reader) (int, map[string]any) {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, path, body)
	// An unbounded HTTP body has length -1. BindBody deliberately ignores a
	// zero-length body, so the injected read failure must reach the decoder.
	if request.ContentLength == 0 {
		request.ContentLength = -1
	}
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid API response: %s", response.Body.String())
	}
	return response.Code, result
}

func assertPluginInputRequestLogsSafe(t *testing.T, response map[string]any, logs []*core.Log, path string, forbidden ...string) {
	t.Helper()
	data, ok := response["data"].(map[string]any)
	if !ok || len(data) != 0 {
		t.Errorf("input/lookup failure returned raw error data: %#v", response)
	}
	for _, entry := range logs {
		if entry.Data["type"] == "request" && entry.Data["url"] == path && entry.Data["details"] != nil {
			t.Errorf("PocketBase request logger retained the raw error cause: %#v", entry.Data["details"])
		}
	}
	encoded, err := json.Marshal(map[string]any{"response": response, "logs": logs})
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range append(forbidden, "json: cannot unmarshal", "Go struct field", "unexpected EOF") {
		if strings.Contains(string(encoded), text) {
			t.Errorf("response/PocketBase _logs leaked %q: %s", text, encoded)
		}
	}
}

func TestPluginMalformedBodiesDoNotReachPocketBaseRequestLogs(t *testing.T) {
	const bodySecret = "body-read-private-token-sentinel"
	for _, route := range []struct {
		name, payload string
		handle        func(*core.RequestEvent) error
	}{
		{"oauth_start", `{"pluginId":["request-payload-secret"]}`, PluginSystemOAuthStart},
		{"oauth_callback", `{"code":["request-payload-secret"]}`, PluginSystemOAuthCallback},
		{"oauth_revoke", `{"instanceId":["request-payload-secret"]}`, PluginSystemOAuthRevoke},
		{"send", `{"pluginId":["request-payload-secret"]}`, PluginSystemTrailSend},
		{"validate", `{"auth":["request-payload-secret"]}`, PluginSystemSessionAuthValidate},
	} {
		t.Run(route.name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			path := "/api/test/" + route.name
			handler := pluginInputRequestLogHandler(t, f.app, path, route.handle)
			enablePluginRequestLogs(t, f.app)
			for i, input := range []struct {
				name string
				body io.Reader
			}{
				{"wrong JSON field type", strings.NewReader(route.payload)},
				{"HTTP body read failure", iotest.ErrReader(errors.New("read request body token=" + bodySecret))},
			} {
				t.Run(input.name, func(t *testing.T) {
					status, response := pluginInputRequestLogPost(t, handler, f.authToken, path, input.body)
					if status != http.StatusBadRequest || response["message"] != "Failed to read request data." {
						t.Errorf("status=%d, expected the fixed body-read failure: %#v", status, response)
					}
					logs := waitForPluginRequestLogs(t, f.app, path, i+1)
					assertPluginInputRequestLogsSafe(t, response, logs, path, "request-payload-secret", bodySecret)
				})
			}
			if f.tokenRequests.Load() != 0 {
				t.Error("a malformed body reached provider I/O")
			}
		})
	}
}

func TestPluginLookupFailureDoesNotRetainRawPocketBaseErrorData(t *testing.T) {
	const missingPluginID = "request-log-missing-plugin"
	for _, route := range []struct {
		name   string
		handle func(*core.RequestEvent) error
		body   func(*oauthMetadataFixture) map[string]any
	}{
		{"oauth_start", PluginSystemOAuthStart, func(f *oauthMetadataFixture) map[string]any {
			return map[string]any{"pluginId": missingPluginID, "instanceId": f.instanceID, "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback"}
		}},
		{"oauth_callback", PluginSystemOAuthCallback, func(f *oauthMetadataFixture) map[string]any {
			return map[string]any{"instanceId": f.instanceID, "code": "unused-code", "state": "unused-state"}
		}},
		{"send", PluginSystemTrailSend, func(f *oauthMetadataFixture) map[string]any {
			return map[string]any{"pluginId": missingPluginID, "trailId": "unused-trail"}
		}},
		{"validate", PluginSystemSessionAuthValidate, func(f *oauthMetadataFixture) map[string]any {
			return map[string]any{"pluginId": missingPluginID, "instanceId": f.instanceID}
		}},
	} {
		t.Run(route.name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			instance, err := f.app.FindRecordById("plugin_instances", f.instanceID)
			if err != nil {
				t.Fatal(err)
			}
			instance.Set("plugin_id", missingPluginID)
			if err := f.app.Save(instance); err != nil {
				t.Fatal(err)
			}
			// Exercise the actual cache/discovery lookup. Its fallback discards
			// cached filesystem/manifest causes, so test the nil-RawData contract
			// directly instead of introducing a directory or permission seam.
			if _, err := pluginsystem.LoadInstalledPlugin(f.app, "", missingPluginID); err == nil {
				t.Fatal("missing-plugin fixture unexpectedly resolves")
			}
			path := "/api/test/" + route.name
			handler := pluginInputRequestLogHandler(t, f.app, path, route.handle)
			enablePluginRequestLogs(t, f.app)
			payload, err := json.Marshal(route.body(f))
			if err != nil {
				t.Fatal(err)
			}
			status, response := pluginInputRequestLogPost(t, handler, f.authToken, path, bytes.NewReader(payload))
			if status != http.StatusBadRequest || response["message"] != "Unknown plugin." {
				t.Errorf("status=%d, expected a fixed unknown-plugin failure: %#v", status, response)
			}
			logs := waitForPluginRequestLogs(t, f.app, path, 1)
			assertPluginInputRequestLogsSafe(t, response, logs, path, f.plugin.Dir, f.app.DataDir())
			if f.tokenRequests.Load() != 0 {
				t.Error("failed plugin lookup reached provider I/O")
			}
		})
	}
}

func TestPluginImplicitInstanceLookupFailureDoesNotReachPocketBaseRequestLogs(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	const missingTable = "instance_lookup_private_token_sentinel"
	instance, err := f.app.FindRecordById("plugin_instances", f.instanceID)
	if err != nil {
		t.Fatal(err)
	}
	ownerID := instance.GetString("user")
	// Keep the real collection metadata, owner filter and DB driver. Replace
	// only this fixture's physical table with a broken SQLite view to provoke
	// an actual lookup error that contains a private identifier.
	for _, statement := range []string{
		`ALTER TABLE "plugin_instances" RENAME TO "plugin_instances_unavailable"`,
		`CREATE VIEW "plugin_instances" AS SELECT * FROM "` + missingTable + `"`,
	} {
		if _, err := f.app.DB().NewQuery(statement).Execute(); err != nil {
			t.Fatal(err)
		}
	}
	_, err = f.app.FindFirstRecordByFilter("plugin_instances", "user={:user} && plugin_id={:plugin_id}", dbx.Params{"user": ownerID, "plugin_id": "strava"})
	if err == nil || !strings.Contains(err.Error(), missingTable) {
		t.Fatalf("broken-view fixture did not produce the expected real SQLite error: %v", err)
	}
	enablePluginRequestLogs(t, f.app)
	path := "/api/plugin-system/oauth/start"
	status, response := interactiveDiagnosticRequest(t, f.handler, f.authToken, path, map[string]any{
		"pluginId": "strava", "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback",
	})
	if status != http.StatusNotFound || response["message"] != "Plugin instance not found." {
		t.Errorf("status=%d, expected the fixed lookup failure: %#v", status, response)
	}
	logs := waitForPluginRequestLogs(t, f.app, path, 1)
	assertPluginInputRequestLogsSafe(t, response, logs, path, missingTable, f.app.DataDir())
	if f.tokenRequests.Load() != 0 {
		t.Error("failed instance lookup reached provider I/O")
	}
}
