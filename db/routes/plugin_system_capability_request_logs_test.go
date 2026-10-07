package routes

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"pocketbase/pluginsystem"
)

func TestSendMissingCapabilityDoesNotPersistRawErrorData(t *testing.T) {
	f := newInteractivePluginFixture(t, nil)
	installed, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", f.manifest.ID)
	if err != nil {
		t.Fatal(err)
	}
	f.manifest.Capabilities = []pluginsystem.CapabilityManifest{{Name: "list_routes", Version: "v1", Export: "list_routes_v1"}}
	installed.Set("manifest", f.manifest)
	if err := f.app.Save(installed); err != nil {
		t.Fatal(err)
	}
	enablePluginRequestLogs(t, f.app)
	const path = "/api/test/send-missing-capability"
	payload, err := json.Marshal(map[string]any{"pluginId": f.manifest.ID, "trailId": "not-reached"})
	if err != nil {
		t.Fatal(err)
	}
	status, body := pluginHostRouteRequest(t, f.app, f.owner, http.MethodPost, path, string(payload), func(e *core.RequestEvent) error {
		// Capability validation must stop before any trail/provider/runtime use.
		return pluginSystemTrailSend(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) {
			t.Error("missing capability reached runtime selection")
			return nil, nil
		})
	})
	if status != http.StatusBadRequest || body["message"] != "Plugin capability is not available." {
		t.Errorf("missing capability changed the static API error: status=%d body=%#v", status, body)
	}
	data, ok := body["data"].(map[string]any)
	if !ok || len(data) != 0 {
		t.Errorf("missing capability returned raw API error data: %#v", body)
	}
	logs := waitForPluginRequestLogs(t, f.app, path, 1)
	requests := 0
	for _, entry := range logs {
		if entry.Data["type"] == "request" && entry.Data["method"] == http.MethodPost && entry.Data["url"] == path {
			requests++
			if entry.Data["error"] != "Plugin capability is not available." {
				t.Errorf("request logger did not retain the static error: %#v", entry.Data)
			}
			// PocketBase writes ApiError.RawData() under "details", before
			// rendering its generic validation messages into the API response.
			if details := entry.Data["details"]; details != nil {
				t.Errorf("capability name/version persisted as raw error details: %#v", details)
			}
			encoded, err := json.Marshal(entry)
			if err != nil {
				t.Fatal(err)
			}
			for _, raw := range []string{`"name"`, `"version"`, "prepare_trail_send", `"v1"`} {
				if strings.Contains(string(encoded), raw) {
					t.Errorf("request log persisted capability name/version %q: %s", raw, encoded)
				}
			}
		}
	}
	if requests != 1 {
		t.Errorf("want one actual PocketBase request log, got %d: %#v", requests, logs)
	}
}
