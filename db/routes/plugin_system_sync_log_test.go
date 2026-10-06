package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

type syncDiagnosticApp struct {
	core.App
	logger *slog.Logger
}

func (a syncDiagnosticApp) Logger() *slog.Logger { return a.logger }

type syncDiagnosticSession struct{}

func (syncDiagnosticSession) Close(context.Context) error { return nil }
func (syncDiagnosticSession) Call(_ context.Context, export string, _ []byte) ([]byte, error) {
	if export == "list" {
		return json.Marshal(pluginSystemListOutput{Items: []pluginsystem.TrailSummary{{Source: pluginsystem.TrailImportSource{Provider: "provider-secret", ExternalID: "https://user:password-secret@host/?token=secret"}}}})
	}
	return nil, pluginsystem.PluginFailure{PluginError: pluginsystem.PluginError{Code: "rate_limited", Message: "body-secret", Operation: pluginsystem.OperationDetail, HTTPStatus: 429}}
}

func TestSyncDetailFailureLogsHostIdentifiersAndOpaqueSource(t *testing.T) {
	var output bytes.Buffer
	app := syncDiagnosticApp{logger: slog.New(slog.NewJSONHandler(&output, nil))}
	collection := core.NewBaseCollection("plugin_instances")
	collection.Fields.Add(&core.TextField{Name: "plugin_id"}, &core.TextField{Name: "user"})
	instance := core.NewRecord(collection)
	instance.Id = "host-instance"
	instance.Set("plugin_id", "host-plugin")
	// No known references/user: the tested batch takes the actual detail failure
	// path without importing a record or needing a provider/database fixture.
	plugin := pluginsystem.LocalPlugin{Manifest: pluginsystem.Manifest{ID: "host-plugin"}}
	sessions := &pluginSyncRuntimeSession{session: syncDiagnosticSession{}}
	result, err := syncPluginCapability(context.Background(), app, nil, sessions, plugin,
		pluginsystem.CapabilityManifest{Name: "list_routes", Export: "list"}, pluginsystem.CapabilityManifest{Name: "get_route_detail", Export: "detail"}, instance, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Skipped != 1 {
		t.Fatalf("detail failure was not exercised: %#v", result)
	}
	if strings.Contains(output.String(), "secret") || strings.Contains(output.String(), "https://") {
		t.Fatalf("guest identifier/error leaked: %s", output.String())
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	var record map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &record); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{"plugin": "host-plugin", "instance": "host-instance", "capability": "list_routes", "code": "rate_limited", "class": "rate_limit", "operation": pluginsystem.OperationDetail, "http_status": float64(429)} {
		if record[key] != want {
			t.Fatalf("lost safe diagnostic %s: %s", key, lines[len(lines)-1])
		}
	}
	if source, ok := record["source_id"].(string); !ok || len(source) != 16 {
		t.Fatalf("missing bounded source correlation: %#v", record)
	}
}
