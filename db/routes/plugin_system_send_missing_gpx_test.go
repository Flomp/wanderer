package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"pocketbase/pluginsystem"
)

func TestSendMissingGPXIsExpectedPrerequisite(t *testing.T) {
	for _, tc := range []struct {
		name, level, code, class, cause   string
		storedFile, removeFile, emptyFile bool
	}{
		{name: "absent field", level: "INFO", code: "invalid_request", class: "configuration", cause: "trail_gpx_missing"},
		{name: "stored file cannot be read", level: "WARN", code: "internal_error", class: "internal", cause: "file_read_failed", storedFile: true, removeFile: true},
		{name: "stored file is empty", level: "WARN", code: "invalid_request", class: "configuration", cause: "trail_gpx_empty", storedFile: true, emptyFile: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newInteractivePluginFixture(t, nil)
			trails := core.NewBaseCollection("trails")
			trails.Fields.Add(&core.BoolField{Name: "public"}, &core.TextField{Name: "name"}, &core.FileField{Name: "gpx", MaxSelect: 1})
			if err := f.app.Save(trails); err != nil {
				t.Fatal(err)
			}
			trail := core.NewRecord(trails)
			trail.Set("public", true)
			trail.Set("name", "guest-trail-name-secret")
			if tc.storedFile {
				contents := []byte(`<gpx version="1.1"></gpx>`)
				file, err := filesystem.NewFileFromBytes(contents, "private_walk_to_my_home.gpx")
				if err != nil {
					t.Fatal(err)
				}
				trail.Set("gpx", file)
			}
			if err := f.app.Save(trail); err != nil {
				t.Fatal(err)
			}
			if tc.removeFile || tc.emptyFile {
				fsys, err := f.app.NewFilesystem()
				if err != nil {
					t.Fatal(err)
				}
				fileKey := trail.BaseFilesPath() + "/" + trail.GetString("gpx")
				if tc.removeFile {
					err = fsys.Delete(fileKey)
				} else {
					// PocketBase rejects empty uploads; simulate a damaged stored
					// file after saving a valid reference, using its real filesystem.
					err = fsys.Upload([]byte{}, fileKey)
				}
				if err != nil {
					t.Fatal(err)
				}
				if err := fsys.Close(); err != nil {
					t.Fatal(err)
				}
			}
			runtimeCalls := 0
			send := func(e *core.RequestEvent) error {
				return pluginSystemTrailSend(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) {
					runtimeCalls++
					return nil, errors.New("unexpected runtime call")
				})
			}
			var output bytes.Buffer
			handler := interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{"/api/test/send": send})
			token, err := f.owner.NewAuthToken()
			if err != nil {
				t.Fatal(err)
			}
			payload := map[string]any{"pluginId": f.manifest.ID, "trailId": trail.Id}
			status, body := interactiveDiagnosticRequest(t, handler, token, "/api/test/send", payload)
			if status != 400 {
				t.Errorf("API status %d: %#v", status, body)
			}
			wantMessage := "Trail has no GPX track."
			if tc.removeFile {
				wantMessage = "Failed to read trail GPX."
			}
			data, _ := body["data"].(map[string]any)
			if body["message"] != wantMessage || len(data) != 0 {
				t.Errorf("host GPX failure was exposed as a plugin failure: %#v", body)
			}
			if runtimeCalls != 0 {
				t.Fatal("missing GPX reached the guest or provider")
			}
			want := map[string]any{"level": tc.level, "action": "send", "owner": f.owner.Id, "plugin": f.manifest.ID, "instance": f.instance.Id, "code": tc.code, "class": tc.class, "operation": pluginsystem.OperationSend, "cause": tc.cause}
			assertInteractiveDiagnostic(t, &output, want, "guest-trail-name-secret", "private_walk_to_my_home", trail.GetString("gpx"))

			// Capture tests do not exercise PocketBase's async request logger.
			// Repeat through its real logger with production log retention enabled.
			enablePluginRequestLogs(t, f.app)
			router, err := apis.NewRouter(f.app)
			if err != nil {
				t.Fatal(err)
			}
			router.POST("/api/test/send", send)
			loggedHandler, err := router.BuildMux()
			if err != nil {
				t.Fatal(err)
			}
			loggedStatus, loggedBody := interactiveDiagnosticRequest(t, loggedHandler, token, "/api/test/send", payload)
			if loggedStatus != status {
				t.Fatalf("logged request changed status: %d %#v", loggedStatus, loggedBody)
			}
			logs := waitForPluginRequestLogs(t, f.app, "/api/test/send", 1)
			if runtimeCalls != 0 {
				t.Fatal("logged missing GPX reached the guest or provider")
			}
			for _, value := range []any{loggedBody, logs} {
				encoded, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				for _, secret := range []string{"guest-trail-name-secret", "private_walk_to_my_home", trail.GetString("gpx"), f.app.DataDir()} {
					if secret != "" && strings.Contains(string(encoded), secret) {
						t.Errorf("response/PocketBase logs leaked %q: %s", secret, encoded)
					}
				}
			}
		})
	}
}
