package routes

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"pocketbase/pluginsystem"
)

func TestInteractiveRuntimeFailureDoesNotReachPocketBaseRequestLogs(t *testing.T) {
	for _, action := range []string{"send", "validate"} {
		t.Run(action, func(t *testing.T) {
			f := newInteractivePluginFixture(t, nil)
			const privateError = "worker /private/credential-secret/guest.wasm failed with token=runtime-token-secret"
			resolve := func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) { return nil, errors.New(privateError) }
			payload := map[string]any{"pluginId": f.manifest.ID, "instanceId": f.instance.Id, "expected_connection_generation": f.instance.GetString("connection_generation")}
			var handlerFunc func(*core.RequestEvent) error
			if action == "validate" {
				handlerFunc = func(e *core.RequestEvent) error { return pluginSystemSessionAuthValidate(e, resolve) }
			} else {
				trails := core.NewBaseCollection("trails")
				trails.Fields.Add(&core.BoolField{Name: "public"}, &core.FileField{Name: "gpx", MaxSelect: 1})
				if err := f.app.Save(trails); err != nil {
					t.Fatal(err)
				}
				file, err := filesystem.NewFileFromBytes([]byte(`<gpx version="1.1"></gpx>`), "private-runtime-trail.gpx")
				if err != nil {
					t.Fatal(err)
				}
				trail := core.NewRecord(trails)
				trail.Set("public", true)
				trail.Set("gpx", file)
				if err := f.app.Save(trail); err != nil {
					t.Fatal(err)
				}
				payload["trailId"] = trail.Id
				handlerFunc = func(e *core.RequestEvent) error { return pluginSystemTrailSend(e, resolve) }
			}
			enablePluginRequestLogs(t, f.app)
			router, err := apis.NewRouter(f.app)
			if err != nil {
				t.Fatal(err)
			}
			path := "/api/test/" + action
			router.POST(path, handlerFunc)
			handler, err := router.BuildMux()
			if err != nil {
				t.Fatal(err)
			}
			token, err := f.owner.NewAuthToken()
			if err != nil {
				t.Fatal(err)
			}
			status, body := interactiveDiagnosticRequest(t, handler, token, path, payload)
			if status != 400 {
				t.Fatalf("status %d: %#v", status, body)
			}
			data, _ := body["data"].(map[string]any)
			if len(data) != 0 || body["message"] != "Failed to initialize plugin runtime." {
				t.Errorf("host runtime initialization was exposed as a plugin failure: %#v", body)
			}
			logs := waitForPluginRequestLogs(t, f.app, path, 1)
			for _, value := range []any{body, logs} {
				encoded, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				for _, secret := range []string{privateError, "/private/", "credential-secret", "runtime-token-secret", "private-runtime-trail"} {
					if strings.Contains(string(encoded), secret) {
						t.Errorf("response/PocketBase logs leaked %q: %s", secret, encoded)
					}
				}
			}
		})
	}
}
