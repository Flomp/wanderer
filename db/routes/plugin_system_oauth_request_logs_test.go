package routes

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/logger"

	"pocketbase/pluginsystem"
)

// TestApp normally disables request logs. These probes retain the original
// PocketBase BatchHandler and its real auxiliary database, with the production
// default retention of five days. A captured diagnostic alone cannot detect a
// raw route error subsequently persisted by the request logger.
func enablePluginRequestLogs(t *testing.T, app core.App) {
	t.Helper()
	handler, ok := app.Logger().Handler().(*logger.BatchHandler)
	if !ok {
		t.Fatalf("persisted request logs require the original PocketBase BatchHandler, got %T", app.Logger().Handler())
	}
	app.Settings().Logs.MaxDays = 5
	app.Settings().Logs.MinLevel = int(slog.LevelInfo)
	handler.SetLevel(slog.LevelInfo)
	if err := handler.WriteAll(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := app.AuxDB().Delete(core.LogsTableName, nil).Execute(); err != nil {
		t.Fatal(err)
	}
}

// PocketBase schedules request logging after the route returns. Flush and read
// the actual _logs table until all expected POST entries exist; a single flush
// can precede that goroutine and give a false assurance that no secret leaked.
// Return every persisted log, including the synchronous plugin diagnostic.
func waitForPluginRequestLogs(t *testing.T, app core.App, path string, count int) []*core.Log {
	t.Helper()
	return waitForPluginRequestLogsForMethod(t, app, http.MethodPost, path, count)
}

func waitForPluginRequestLogsForMethod(t *testing.T, app core.App, method, path string, count int) []*core.Log {
	t.Helper()
	handler, ok := app.Logger().Handler().(*logger.BatchHandler)
	if !ok {
		t.Fatalf("persisted request logs require the original PocketBase BatchHandler, got %T", app.Logger().Handler())
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err := handler.WriteAll(context.Background()); err != nil {
			t.Fatal(err)
		}
		var logs []*core.Log
		if err := app.LogQuery().OrderBy("created", "id").All(&logs); err != nil {
			t.Fatal(err)
		}
		requests := 0
		for _, entry := range logs {
			if entry.Data["type"] == "request" && entry.Data["method"] == method && entry.Data["url"] == path {
				requests++
			}
		}
		if requests >= count {
			return logs
		}
		if time.Now().After(deadline) {
			t.Fatalf("persisted request log for %s %s missing: got %d, want %d", method, path, requests, count)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestOAuthSaveFailuresDoNotLeakIntoPocketBaseRequestLogs(t *testing.T) {
	const secret = "oauth-save-failure-secret-sentinel"
	for _, action := range []string{"start", "callback", "revoke"} {
		t.Run(action, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			path := "/api/plugin-system/oauth/" + action
			body := map[string]any{"instanceId": f.instanceID}
			switch action {
			case "start":
				body["pluginId"] = "strava"
				body["redirectUri"] = "https://wanderer.example/settings/plugins/oauth/callback"
			case "callback":
				start := f.request(t, http.MethodPost, "/api/plugin-system/oauth/start", map[string]any{
					"instanceId": f.instanceID, "pluginId": "strava",
					"redirectUri": "https://wanderer.example/settings/plugins/oauth/callback",
				})
				body["state"] = start["state"]
				body["code"] = "test-authorization-code"
			}
			enablePluginRequestLogs(t, f.app)
			before, err := f.app.FindRecordById("plugin_instances", f.instanceID)
			if err != nil {
				t.Fatal(err)
			}
			beforeAuth := pluginsystem.JSONMapFromRecord(before, "auth")
			f.app.OnRecordUpdate("plugin_instances").BindFunc(func(e *core.RecordEvent) error {
				return errors.New("SQL UPDATE auth=" + secret + " provider=https://oauth.invalid/?token=" + secret)
			})

			status, result := interactiveDiagnosticRequest(t, f.handler, f.authToken, path, body)
			if status != http.StatusBadRequest {
				t.Errorf("status=%d, want 400: %#v", status, result)
			}
			response, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(response), secret) {
				t.Errorf("API response leaked the save cause: %s", response)
			}
			data, _ := result["data"].(map[string]any)
			if len(data) != 0 || result["message"] != "Failed to save plugin connection." {
				t.Errorf("host save failure was exposed as a plugin failure: %s", response)
			}

			logs := waitForPluginRequestLogs(t, f.app, path, 1)
			persisted, err := json.Marshal(logs)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(persisted), secret) {
				t.Errorf("PocketBase _logs persisted the raw save cause: %s", persisted)
			}
			diagnostics, requests := 0, 0
			for _, entry := range logs {
				if entry.Message == "plugin operation failed" {
					diagnostics++
					for key, want := range map[string]any{
						"action": "oauth_" + action, "operation": "oauth_" + action,
						"code": "internal_error", "class": "internal", "cause": "database_save",
					} {
						if entry.Data[key] != want {
							t.Errorf("persisted diagnostic %s=%#v, want %#v", key, entry.Data[key], want)
						}
					}
				}
				if entry.Data["type"] == "request" && entry.Data["url"] == path {
					requests++
					if entry.Data["error"] != "Failed to save plugin connection." {
						t.Errorf("request logger did not receive the safe API error: %#v", entry.Data)
					}
				}
			}
			if diagnostics != 1 || requests != 1 {
				t.Errorf("want one plugin diagnostic and one actual request log, got %d/%d: %s", diagnostics, requests, persisted)
			}
			after, err := f.app.FindRecordById("plugin_instances", f.instanceID)
			if err != nil {
				t.Fatal(err)
			}
			if before.GetString("status") != after.GetString("status") || !reflect.DeepEqual(beforeAuth, pluginsystem.JSONMapFromRecord(after, "auth")) {
				t.Error("failed save changed the persisted OAuth state")
			}
			wantRequests := int32(0)
			if action == "callback" {
				wantRequests = 1
			}
			if got := f.tokenRequests.Load(); got != wantRequests {
				t.Errorf("provider requests=%d, want %d", got, wantRequests)
			}
		})
	}
}

func TestOAuthInvalidRedirectDoesNotLeakIntoPocketBaseRequestLogs(t *testing.T) {
	const secret = "oauth-invalid-redirect-secret-sentinel"
	f := newOAuthMetadataFixture(t)
	enablePluginRequestLogs(t, f.app)
	path := "/api/plugin-system/oauth/start"
	status, response := interactiveDiagnosticRequest(t, f.handler, f.authToken, path, map[string]any{
		"pluginId": "strava", "instanceId": f.instanceID,
		"redirectUri": "https://wanderer.example/%zz/settings/plugins/oauth/callback?token=" + secret,
	})
	if status != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400: %#v", status, response)
	}
	logs := waitForPluginRequestLogs(t, f.app, path, 1)
	encoded, err := json.Marshal(map[string]any{"response": response, "logs": logs})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), secret) {
		t.Errorf("invalid redirect leaked through API/request-log raw error data: %s", encoded)
	}
	if f.tokenRequests.Load() != 0 {
		t.Error("invalid redirect contacted the provider")
	}
}
