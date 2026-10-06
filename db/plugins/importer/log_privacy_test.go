package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
	"pocketbase/pluginsystem"
	"pocketbase/util"
)

type mediaDiagnosticApp struct {
	core.App
	logger *slog.Logger
}

func (a mediaDiagnosticApp) Logger() *slog.Logger { return a.logger }

func TestPhotoDiagnosticsPreserveSafeFailureAndHostIdentifiers(t *testing.T) {
	for _, tc := range []struct {
		err                    error
		code, class, operation string
		status                 int
	}{
		{&url.Error{Op: "Get", URL: "https://user:password-secret@provider.test/?token=secret", Err: errors.New("provider-body-secret")}, "provider_unavailable", "transient", pluginsystem.OperationHostHTTP, 0},
		{util.PluginMediaHTTPStatusError{StatusCode: 503}, "provider_unavailable", "transient", pluginsystem.OperationHostHTTP, 503},
	} {
		var output bytes.Buffer
		app := mediaDiagnosticApp{logger: slog.New(slog.NewJSONHandler(&output, nil))}
		ctx := pluginMediaLogContext{PluginID: "host-plugin", TrailID: "local-trail-id", SourceID: pluginsystem.SourceDiagnosticID("provider-secret", "external-secret")}
		photo := pluginsystem.Photo{Filename: "filename-secret.jpg", ExternalID: "photo-id-secret", Source: pluginsystem.MediaSource{Type: "source-type-secret", URL: "https://query-secret.invalid"}}
		logSkippedPluginPhoto(app, "photo skipped", ctx, photo, tc.err)
		var record map[string]any
		if err := json.Unmarshal(output.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		for key, want := range map[string]any{"plugin": "host-plugin", "trail": "local-trail-id", "code": tc.code, "class": tc.class, "operation": tc.operation, "http_status": float64(tc.status), "cause": "media_fetch_failed", "source_type": "unknown"} {
			if record[key] != want {
				t.Fatalf("lost safe field %s: %s", key, output.String())
			}
		}
		if strings.Contains(output.String(), "secret") || strings.Contains(output.String(), "provider.test") || strings.Contains(output.String(), "https://") {
			t.Fatalf("photo data/error leaked: %s", output.String())
		}
	}
}

func TestImportTrailAndWaypointPhotoFailuresDoNotLogGuestNames(t *testing.T) {
	realApp, err := pbtests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer realApp.Cleanup()
	for name, fields := range map[string][]string{
		"activitypub_actors":       {"user"},
		"trails":                   {"author", "name", "description"},
		"waypoints":                {"trail", "author", "name", "description"},
		"trail_external_reference": {"trail", "user", "provider", "external_id", "plugin_id", "provider_category"},
	} {
		collection := core.NewBaseCollection(name)
		for _, field := range fields {
			collection.Fields.Add(&core.TextField{Name: field})
		}
		if name == "trails" {
			collection.Fields.Add(&core.FileField{Name: "gpx", MaxSelect: 1})
		}
		if name == "waypoints" {
			collection.Fields.Add(&core.NumberField{Name: "lat"}, &core.NumberField{Name: "lon"})
		}
		if name == "trail_external_reference" {
			collection.Fields.Add(&core.DateField{Name: "provider_category_checked_at"}, &core.AutodateField{Name: "created", OnCreate: true})
		}
		if err := realApp.Save(collection); err != nil {
			t.Fatal(err)
		}
	}
	actors, _ := realApp.FindCollectionByNameOrId("activitypub_actors")
	actor := core.NewRecord(actors)
	actor.Set("user", "host-user")
	if err := realApp.Save(actor); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	app := mediaDiagnosticApp{App: realApp, logger: slog.New(slog.NewJSONHandler(&output, nil))}
	photo := pluginsystem.Photo{Filename: "filename-secret.jpg", ExternalID: "photo-id-secret", Source: pluginsystem.MediaSource{Type: "url", URL: "https://[invalid?token=secret"}}
	item := pluginsystem.TrailImport{
		Name: "trail-name-secret", Source: pluginsystem.TrailImportSource{Provider: "provider-secret", ExternalID: "trail-id-secret"}, Track: gpxTrack(),
		Photos: []pluginsystem.Photo{photo}, Waypoints: []pluginsystem.Waypoint{{Name: "waypoint-name-secret", Lat: 46, Lon: 8, Photos: []pluginsystem.Photo{photo}}},
	}
	result, err := ImportTrail(context.Background(), app, item, Options{UserID: "host-user", ActorID: actor.Id, Manifest: pluginsystem.Manifest{ID: "host-plugin"}})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected actual trail+waypoint photo failures: %s", output.String())
	}
	if strings.Contains(output.String(), "secret") || strings.Contains(output.String(), "https://") {
		t.Fatalf("guest metadata leaked during import: %s", output.String())
	}
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		if record["code"] != "invalid_request" || record["class"] != "configuration" || record["cause"] != "photo_source_invalid" {
			t.Fatalf("malformed photo source was treated as provider outage during import: %s", line)
		}
	}
	var waypointLog map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &waypointLog); err != nil {
		t.Fatal(err)
	}
	if waypointLog["plugin"] != "host-plugin" || waypointLog["trail"] != result.TrailID || waypointLog["operation"] != pluginsystem.OperationHostHTTP {
		t.Fatalf("local identifiers/operation missing from waypoint diagnostics: %s", lines[1])
	}
}

func TestPhotoFileRejectsLocalURLBeforeFetch(t *testing.T) {
	for _, rawURL := range []string{
		"https://[invalid?token=secret",
		"https://user:password-secret@example.test/photo?token=secret",
		"ftp://example.test/photo?token=secret",
		"https://",
		"",
	} {
		t.Run(rawURL, func(t *testing.T) {
			// A canceled transport context still returns the local source error:
			// validation must happen before network/fetch or retry machinery.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			photo := pluginsystem.Photo{Source: pluginsystem.MediaSource{Type: "url", URL: rawURL}}
			file, size, err := photoFile(ctx, photo, Options{}, util.DefaultPluginMediaMaxBytes, nil)
			safe := pluginsystem.PublicPluginError(err)
			if file != nil || size != 0 || err == nil || safe.Code != "invalid_request" || safe.Operation != pluginsystem.OperationHostHTTP || safe.HTTPStatus != 0 {
				t.Fatalf("local malformed photo URL became provider failure: file=%v size=%d error=%#v", file, size, safe)
			}
			var output bytes.Buffer
			app := mediaDiagnosticApp{logger: slog.New(slog.NewJSONHandler(&output, nil))}
			logSkippedPluginPhoto(app, "photo skipped", pluginMediaLogContext{PluginID: "host-plugin"}, photo, err)
			var record map[string]any
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if record["code"] != "invalid_request" || record["class"] != "configuration" || record["cause"] != "photo_source_invalid" {
				t.Fatalf("local URL failure lost its classification: %s", output.String())
			}
			if strings.Contains(output.String(), "secret") || strings.Contains(output.String(), "https://") || strings.Contains(err.Error(), "secret") {
				t.Fatalf("local source data leaked: %s / %v", output.String(), err)
			}
		})
	}
}
