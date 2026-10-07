package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

type guardedSyncSession struct {
	listCalls   atomic.Int32
	detailCalls atomic.Int32
	arrived     chan struct{}
	resume      chan struct{}
}

func (s *guardedSyncSession) Close(context.Context) error { return nil }
func (s *guardedSyncSession) Call(_ context.Context, export string, _ []byte) ([]byte, error) {
	if export == "list" {
		if s.listCalls.Add(1) == 1 {
			return json.Marshal(pluginSystemListOutput{Items: []pluginsystem.TrailSummary{{Source: pluginsystem.TrailImportSource{Provider: "strava", ExternalID: "item-one"}}, {Source: pluginsystem.TrailImportSource{Provider: "strava", ExternalID: "item-two"}}}, HasMore: true})
		}
		return json.Marshal(pluginSystemListOutput{})
	}
	if s.detailCalls.Add(1) == 1 {
		close(s.arrived)
		<-s.resume
	}
	return json.Marshal(pluginSystemDetailOutput{Item: pluginsystem.TrailImport{Source: pluginsystem.TrailImportSource{Provider: "strava", ExternalID: "item-one"}}})
}

func prepareSyncReferenceCollections(t *testing.T, f *oauthMetadataFixture, existing bool) *core.Record {
	t.Helper()
	trails := core.NewBaseCollection("trails")
	trails.Fields.Add(&core.TextField{Name: "user"})
	if err := f.app.Save(trails); err != nil {
		t.Fatal(err)
	}
	refs := core.NewBaseCollection("trail_external_reference")
	refs.Fields.Add(&core.TextField{Name: "user"}, &core.TextField{Name: "provider"}, &core.TextField{Name: "external_id"}, &core.TextField{Name: "trail"}, &core.TextField{Name: "provider_category"}, &core.DateField{Name: "provider_category_checked_at"})
	if err := f.app.Save(refs); err != nil {
		t.Fatal(err)
	}
	if !existing {
		return nil
	}
	instance, _ := f.storedAuth(t)
	var first *core.Record
	for _, id := range []string{"item-one", "item-two"} {
		trail := core.NewRecord(trails)
		trail.Set("user", instance.GetString("user"))
		if err := f.app.Save(trail); err != nil {
			t.Fatal(err)
		}
		ref := core.NewRecord(refs)
		ref.Set("user", instance.GetString("user"))
		ref.Set("provider", "strava")
		ref.Set("external_id", id)
		ref.Set("trail", trail.Id)
		if err := f.app.Save(ref); err != nil {
			t.Fatal(err)
		}
		if first == nil {
			first = ref
		}
	}
	return first
}

func TestSyncChecksFreshConnectionBeforeImportAndNextItem(t *testing.T) {
	for _, mutation := range []string{"disable", "revoke", "delete", "reconnect"} {
		t.Run(mutation, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			prepareSyncReferenceCollections(t, f, false)
			instance, auth := f.storedAuth(t)
			flow := startConnectionFlow(t, f, false)
			session := &guardedSyncSession{arrived: make(chan struct{}), resume: make(chan struct{})}
			cleanupConnectionBarrier(t, session.resume)
			result := make(chan error, 1)
			go func() {
				_, err := syncPluginCapability(context.Background(), f.app, nil, &pluginSyncRuntimeSession{session: session}, f.plugin, pluginsystem.CapabilityManifest{Name: "list_routes", Export: "list"}, pluginsystem.CapabilityManifest{Name: "get_route_detail", Export: "detail"}, instance, nil, auth, nil, nil, false, false)
				result <- err
			}()
			waitConnectionBarrier(t, session.arrived)
			switch mutation {
			case "disable":
				f.request(t, http.MethodPatch, "/api/plugins/instances/"+f.instanceID+"/settings", map[string]any{"enabled": false})
			case "revoke":
				f.request(t, http.MethodPost, "/api/plugin-system/oauth/revoke", map[string]any{"instanceId": f.instanceID})
			case "delete":
				if err := f.app.Delete(instance); err != nil {
					t.Fatal(err)
				}
			case "reconnect":
				f.request(t, http.MethodPost, "/api/plugin-system/oauth/callback", map[string]any{"instanceId": f.instanceID, "state": flow["state"], "code": "test-authorization-code"})
			}
			var afterMutation map[string]any
			if mutation != "delete" {
				fresh, _ := f.storedAuth(t)
				afterMutation = fresh.PublicExport()
			}
			close(session.resume)
			select {
			case err := <-result:
				if !errors.Is(err, pluginsystem.ErrConnectionObsolete) {
					t.Fatalf("obsolete sync: %v", err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("sync stuck")
			}
			if session.listCalls.Load() != 1 || session.detailCalls.Load() != 1 {
				t.Fatal("obsolete sync started another batch/item")
			}
			trails, err := f.app.FindAllRecords("trails")
			if err != nil || len(trails) != 0 {
				t.Fatal("obsolete detail reached import")
			}
			if err := completePluginSync(f.app, instance, time.Now()); !errors.Is(err, pluginsystem.ErrConnectionObsolete) {
				t.Fatalf("obsolete completion: %v", err)
			}
			setPluginInstanceStatusForError(f.app, instance, pluginsystem.Failure("auth_failed", pluginsystem.OperationLogin))
			if mutation != "delete" {
				fresh, _ := f.storedAuth(t)
				if !reflect.DeepEqual(afterMutation, fresh.PublicExport()) || !fresh.GetDateTime("last_sync_at").IsZero() {
					t.Fatal("obsolete sync wrote status/completion")
				}
			} else if fresh, _ := f.app.FindRecordById("plugin_instances", f.instanceID); fresh != nil {
				t.Fatal("obsolete sync recreated deleted instance")
			}
		})
	}
}

func TestOAuthStartPreservesRunningSyncAndBackfillWrites(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	ref := prepareSyncReferenceCollections(t, f, true)
	instance, auth := f.storedAuth(t)
	session := &guardedSyncSession{arrived: make(chan struct{}), resume: make(chan struct{})}
	cleanupConnectionBarrier(t, session.resume)
	result := make(chan error, 1)
	go func() {
		_, err := syncPluginCapability(context.Background(), f.app, nil, &pluginSyncRuntimeSession{session: session}, f.plugin, pluginsystem.CapabilityManifest{Name: "list_routes", Export: "list"}, pluginsystem.CapabilityManifest{Name: "get_route_detail", Export: "detail"}, instance, nil, auth, nil, map[string]any{"categoryMapping": map[string]any{"Run": "Running"}}, false, false)
		result <- err
	}()
	waitConnectionBarrier(t, session.arrived)
	flow := startConnectionFlow(t, f, true)
	close(session.resume)
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("sync stuck")
	}
	if session.listCalls.Load() != 2 || session.detailCalls.Load() != 2 {
		t.Fatal("ordinary start invalidated existing sync")
	}
	freshRef, err := f.app.FindRecordById("trail_external_reference", ref.Id)
	if err != nil || freshRef.GetDateTime("provider_category_checked_at").IsZero() {
		t.Fatal("valid backfill did not complete")
	}
	if err := completePluginSync(f.app, instance, time.Now()); err != nil {
		t.Fatal(err)
	}
	fresh, stored := f.storedAuth(t)
	if stored["oauthState"] != flow["state"] || fresh.GetDateTime("last_sync_at").IsZero() {
		t.Fatal("completion lost flow or completion time")
	}
}

func TestSyncCompletionKeepsConcurrentStateErrorAndMonotoneTime(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	instance, _ := f.storedAuth(t)
	instance.Set("state", map[string]any{"oldCursor": "observed"})
	if err := f.app.Save(instance); err != nil {
		t.Fatal(err)
	}
	snapshot := pluginsystem.CaptureConnection(instance)
	newer := time.Now().Add(time.Hour)
	_, err := pluginsystem.PatchConnection(f.app, snapshot, true, func(fresh *core.Record) error {
		fresh.Set("state", map[string]any{"oldCursor": "observed", "concurrent": "keep"})
		fresh.Set("last_error", map[string]any{"code": "rate_limited", "authContext": snapshot.AuthContext, "connectionGeneration": snapshot.Generation})
		fresh.Set("retry_not_before", newer)
		fresh.Set("last_sync_at", newer)
		fresh.Set("status", "rate_limited")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := completePluginSync(f.app, instance, time.Now()); err != nil {
		t.Fatal(err)
	}
	fresh, _ := f.storedAuth(t)
	if !reflect.DeepEqual(pluginsystem.JSONMapFromRecord(fresh, "state"), map[string]any{"concurrent": "keep"}) || fresh.GetString("status") != "rate_limited" || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) == 0 || fresh.GetDateTime("last_sync_at").Time().Before(newer.Add(-time.Millisecond)) {
		t.Fatal("completion overwrote concurrent fields or moved time backwards")
	}
}

func TestRefreshThenFullSyncClearsOnlyObservedPriorTokenDiagnosis(t *testing.T) {
	for _, concurrentRotation := range []bool{false, true} {
		name := "unchanged_current_token"
		if concurrentRotation {
			name = "new_current_token"
		}
		t.Run(name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			record, _ := f.storedAuth(t)
			c := pluginsystem.CaptureConnection(record)
			oldT := record.GetString("oauth_token_revision")
			diagnosis := map[string]any{"code": "unauthorized", "operation": pluginsystem.OperationList, "authContext": c.AuthContext, "connectionGeneration": c.Generation, "oauthTokenRevision": oldT}
			record.Set("last_error", diagnosis)
			record.Set("status", "needs_reauth")
			record.Set("retry_not_before", time.Now().Add(time.Minute))
			if err := f.app.Save(record); err != nil {
				t.Fatal(err)
			}
			record, auth := f.storedAuth(t)
			if _, err := pluginsystem.RefreshOAuthToken(context.Background(), f.app, f.plugin, record, auth, c.AuthContext); err != nil {
				t.Fatal(err)
			}
			if record.GetString("oauth_token_revision") == oldT {
				t.Fatal("refresh did not update operation T")
			}
			if concurrentRotation {
				_, err := pluginsystem.PatchConnection(f.app, c, false, func(fresh *core.Record) error {
					fresh.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := completePluginSync(f.app, record, time.Now()); err != nil {
				t.Fatal(err)
			}
			fresh, _ := f.storedAuth(t)
			if concurrentRotation {
				if len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) == 0 || fresh.GetString("status") != "needs_reauth" || fresh.GetDateTime("retry_not_before").IsZero() {
					t.Fatal("completion cleared diagnosis on an unobserved token revision")
				}
			} else {
				if len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || fresh.GetString("status") != "configured" || !fresh.GetDateTime("retry_not_before").IsZero() {
					t.Fatal("successful full sync retained unchanged prior-token failure")
				}
			}
		})
	}
}
