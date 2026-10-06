package routes

import (
	"bytes"
	"errors"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"pocketbase/pluginsystem"
)

func TestSendMissingGPXIsExpectedPrerequisite(t *testing.T) {
	for _, missingStoredFile := range []bool{false, true} {
		name := "absent file"
		if missingStoredFile {
			name = "stored file cannot be read"
		}
		t.Run(name, func(t *testing.T) {
			f := newInteractivePluginFixture(t, nil)
			trails := core.NewBaseCollection("trails")
			trails.Fields.Add(&core.BoolField{Name: "public"}, &core.TextField{Name: "name"}, &core.FileField{Name: "gpx", MaxSelect: 1})
			if err := f.app.Save(trails); err != nil {
				t.Fatal(err)
			}
			trail := core.NewRecord(trails)
			trail.Set("public", true)
			trail.Set("name", "guest-trail-name-secret")
			if missingStoredFile {
				file, err := filesystem.NewFileFromBytes([]byte(`<gpx version="1.1"></gpx>`), "file-name-secret.gpx")
				if err != nil {
					t.Fatal(err)
				}
				trail.Set("gpx", file)
			}
			if err := f.app.Save(trail); err != nil {
				t.Fatal(err)
			}
			if missingStoredFile {
				fsys, err := f.app.NewFilesystem()
				if err != nil {
					t.Fatal(err)
				}
				if err := fsys.Delete(trail.BaseFilesPath() + "/" + trail.GetString("gpx")); err != nil {
					t.Fatal(err)
				}
				if err := fsys.Close(); err != nil {
					t.Fatal(err)
				}
			}
			var output bytes.Buffer
			runtimeCalls := 0
			handler := interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{"/api/test/send": func(e *core.RequestEvent) error {
				return pluginSystemTrailSend(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) {
					runtimeCalls++
					return nil, errors.New("unexpected runtime call")
				})
			}})
			token, err := f.owner.NewAuthToken()
			if err != nil {
				t.Fatal(err)
			}
			status, body := interactiveDiagnosticRequest(t, handler, token, "/api/test/send", map[string]any{"pluginId": f.manifest.ID, "trailId": trail.Id})
			if !missingStoredFile && (status != 400 || body["message"] != "Trail has no GPX track.") {
				t.Fatalf("API response changed: status %d %#v", status, body)
			}
			if missingStoredFile && status == 200 {
				t.Fatal("unreadable file unexpectedly sent")
			}
			if runtimeCalls != 0 {
				t.Fatal("missing GPX reached the guest or provider")
			}
			want := map[string]any{"level": "INFO", "action": "send", "owner": f.owner.Id, "plugin": f.manifest.ID, "instance": f.instance.Id, "code": "invalid_request", "class": "configuration", "operation": pluginsystem.OperationSend, "cause": "trail_gpx_missing"}
			if missingStoredFile {
				want["level"], want["code"], want["class"], want["cause"] = "WARN", "internal_error", "internal", "file_read_failed"
			}
			assertInteractiveDiagnostic(t, &output, want, "guest-trail-name-secret", "file-name-secret.gpx", trail.GetString("gpx"))
		})
	}
}
