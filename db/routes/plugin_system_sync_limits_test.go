package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"pocketbase/pluginsystem"

	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
)

func TestProviderCategoryBackfillQueriesAboveFilterLimit(t *testing.T) {
	app, err := pbtests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	collection := core.NewBaseCollection("trail_external_reference")
	collection.Fields.Add(&core.TextField{Name: "user"}, &core.TextField{Name: "provider"}, &core.TextField{Name: "external_id"},
		&core.TextField{Name: "provider_category"}, &core.DateField{Name: "provider_category_checked_at"})
	if err := app.Save(collection); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, 227)
	for i := range 225 {
		id := fmt.Sprintf("remote-%d", i)
		ids = append(ids, id)
		record := core.NewRecord(collection)
		record.Load(map[string]any{"user": "user", "provider": "provider", "external_id": id})
		if i == 0 {
			record.Set("provider_category", "known")
		}
		if i == 1 {
			record.Set("provider_category_checked_at", "2026-10-12 10:00:00.000Z")
		}
		if err := app.Save(record); err != nil {
			t.Fatal(err)
		}
	}
	ids = append(ids, ids[0], "")
	candidates, err := providerCategoryBackfillCandidatesForSync(app, "user", "provider", ids, 300)
	if err != nil || len(candidates) != 223 || candidates["remote-0"] != nil || candidates["remote-1"] != nil {
		t.Fatalf("large backfill query: %d candidates, error=%v", len(candidates), err)
	}
	limited, err := providerCategoryBackfillCandidatesForSync(app, "user", "provider", ids, 10)
	if err != nil || len(limited) != 10 {
		t.Fatalf("global backfill limit: %d candidates, error=%v", len(limited), err)
	}
}

func TestSyncPluginRejectsOversizedListBeforeAnyStorageAccess(t *testing.T) {
	items := make([]pluginsystem.TrailSummary, defaultPluginSyncBatchLimit+1)
	output, err := json.Marshal(pluginSystemListOutput{Items: items})
	if err != nil {
		t.Fatal(err)
	}
	session := &pluginSyncRuntimeSession{session: syncLimitTestSession{call: func(context.Context, string, []byte) ([]byte, error) {
		return output, nil
	}}}
	instance := core.NewRecord(core.NewBaseCollection("installed_plugins"))
	// A nil app makes any lookup or write before validating maxItems fail.
	_, err = syncPluginCapability(context.Background(), nil, nil, session, pluginsystem.LocalPlugin{},
		pluginsystem.CapabilityManifest{Export: "list_routes_v1"}, pluginsystem.CapabilityManifest{}, instance, nil,
		nil, nil, nil, false, false)
	if err == nil || !strings.Contains(err.Error(), "exceeding maxItems 50") {
		t.Fatalf("oversized list accepted: %v", err)
	}
}

func TestSyncItemLimitHasOwnInstanceCategory(t *testing.T) {
	items := make([]pluginsystem.TrailSummary, defaultPluginSyncBatchLimit+1)
	data, _ := json.Marshal(pluginSystemListOutput{Items: items})
	session := &pluginSyncRuntimeSession{session: syncLimitTestSession{call: func(context.Context, string, []byte) ([]byte, error) {
		return data, nil
	}}}
	instance := core.NewRecord(core.NewBaseCollection("plugin_instances"))
	_, err := syncPluginCapability(t.Context(), nil, nil, session, pluginsystem.LocalPlugin{}, pluginsystem.CapabilityManifest{Export: "list_routes_v1"}, pluginsystem.CapabilityManifest{}, instance, nil, nil, nil, nil, false, false)
	var limit pluginsystem.SyncItemLimitError
	if !errors.As(err, &limit) || limit.Returned != 51 || limit.MaxItems != 50 {
		t.Fatalf("lost protocol limit: %v", err)
	}
	update := pluginsystem.InstanceStatusForError(err, time.Now())
	if update.Code != "sync_item_limit_exceeded" || update.Status != "error" || update.RetryNotBefore != nil || update.Message != pluginsystem.SyncItemLimitMessage {
		t.Fatalf("item limit became outage: %#v", update)
	}
}

// This fake only implements the existing trail-sync runtime, without asset fixtures.
type syncLimitTestSession struct {
	call func(context.Context, string, []byte) ([]byte, error)
}

func (s syncLimitTestSession) Call(ctx context.Context, export string, input []byte) ([]byte, error) {
	return s.call(ctx, export, input)
}

func (syncLimitTestSession) Close(context.Context) error { return nil }

func TestSyncPluginAcceptsListAtItemLimit(t *testing.T) {
	app, err := pbtests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	output, err := json.Marshal(pluginSystemListOutput{Items: make([]pluginsystem.TrailSummary, defaultPluginSyncBatchLimit)})
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	session := &pluginSyncRuntimeSession{session: syncLimitTestSession{call: func(_ context.Context, _ string, data []byte) ([]byte, error) {
		calls++
		var input pluginSystemListInput
		if err := json.Unmarshal(data, &input); err != nil {
			t.Fatal(err)
		}
		if input.Limits.MaxItems != defaultPluginSyncBatchLimit {
			t.Fatalf("maxItems = %d; want %d", input.Limits.MaxItems, defaultPluginSyncBatchLimit)
		}
		return output, nil
	}}}
	instance := core.NewRecord(core.NewBaseCollection("plugin_instances"))
	result, err := syncPluginCapability(t.Context(), app, nil, session, pluginsystem.LocalPlugin{},
		pluginsystem.CapabilityManifest{Export: "list_routes_v1"}, pluginsystem.CapabilityManifest{}, instance, nil,
		nil, nil, nil, false, false)
	if err != nil || result == nil || calls != 1 {
		t.Fatalf("list at limit rejected: result=%#v, calls=%d, error=%v", result, calls, err)
	}
}
