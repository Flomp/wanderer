package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

type diagnosisListSession struct {
	arrived, resume chan struct{}
	failure         error
	calls           atomic.Int32
}

func (*diagnosisListSession) Close(context.Context) error { return nil }
func (s *diagnosisListSession) Call(context.Context, string, []byte) ([]byte, error) {
	s.calls.Add(1)
	close(s.arrived)
	<-s.resume
	if s.failure != nil {
		return nil, s.failure
	}
	return json.Marshal(pluginSystemListOutput{})
}

func TestStatusWriterRejectsOldTokenAndOwnsOnlyDiagnosisTuple(t *testing.T) {
	for _, rotate := range []bool{false, true} {
		name := "same_token"
		if rotate {
			name = "concurrent_token_rotation"
		}
		t.Run(name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			observed := seedRuntimeSideFields(t, f)
			_, auth := f.storedAuth(t)
			session := &diagnosisListSession{arrived: make(chan struct{}), resume: make(chan struct{}), failure: pluginsystem.PluginFailure{PluginError: pluginsystem.PluginError{Code: "rate_limited", Operation: pluginsystem.OperationList, HTTPStatus: 429, RetryAfterSeconds: runtimeIntPointer(120)}}}
			cleanupConnectionBarrier(t, session.resume)
			result := make(chan error, 1)
			go func() {
				_, err := syncPluginCapability(context.Background(), f.app, nil, &pluginSyncRuntimeSession{session: session}, f.plugin, pluginsystem.CapabilityManifest{Name: "list_routes", Export: "list"}, pluginsystem.CapabilityManifest{Name: "get_route_detail", Export: "detail"}, observed, nil, auth, nil, nil, false, false)
				setPluginInstanceStatusForError(f.app, observed, err)
				result <- err
			}()
			waitConnectionBarrier(t, session.arrived)
			startConnectionFlow(t, f, true)
			_, err := pluginsystem.PatchConnection(f.app, pluginsystem.CaptureConnection(observed), true, func(fresh *core.Record) error {
				fresh.Set("config", map[string]any{"host": map[string]any{"completed": false}})
				fresh.Set("state", map[string]any{"newCursor": "must-survive-status"})
				if rotate {
					fresh.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			before, _ := f.storedAuth(t)
			close(session.resume)
			select {
			case err := <-result:
				if err == nil || pluginsystem.PublicPluginError(err).Code != "rate_limited" {
					t.Fatal("list failure did not reach status writer")
				}
			case <-time.After(10 * time.Second):
				t.Fatal("list failure did not finish")
			}
			if session.calls.Load() != 1 {
				t.Fatal("failed list was replayed")
			}
			fresh, _ := f.storedAuth(t)
			if rotate {
				assertRuntimeRecordChanges(t, before, fresh)
			} else {
				assertRuntimeRecordChanges(t, before, fresh, "status", "last_error", "retry_not_before")
				diagnosis := pluginsystem.JSONMapFromRecord(fresh, "last_error")
				c := pluginsystem.CaptureConnection(observed)
				if fresh.GetString("status") != "rate_limited" || fresh.GetDateTime("retry_not_before").IsZero() || diagnosis["authContext"] != c.AuthContext || diagnosis["connectionGeneration"] != c.Generation || diagnosis["oauthTokenRevision"] != observed.GetString("oauth_token_revision") || diagnosis["operation"] != pluginsystem.OperationList || diagnosis["code"] != "rate_limited" || diagnosis["httpStatus"] != float64(429) {
					t.Fatal("status writer lost its bound diagnosis tuple")
				}
			}
		})
	}
}

func runtimeIntPointer(value int) *int { return &value }

func TestSyncCompletionClearsOnlyObservedStateAndDiagnosisTuple(t *testing.T) {
	for _, mutation := range []string{"unchanged", "new_state", "changed_observed_state", "changed_error", "changed_backoff", "changed_token", "foreign_context", "foreign_generation", "historical"} {
		t.Run(mutation, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			observed := seedRuntimeSideFields(t, f)
			c := pluginsystem.CaptureConnection(observed)
			diagnosis := map[string]any{"code": "rate_limited", "operation": pluginsystem.OperationList, "authContext": c.AuthContext, "connectionGeneration": c.Generation, "oauthTokenRevision": observed.GetString("oauth_token_revision")}
			if mutation == "historical" {
				diagnosis = map[string]any{"code": "provider_unavailable", "operation": pluginsystem.OperationList}
			}
			if mutation == "foreign_context" {
				diagnosis["authContext"] = "other-context"
			}
			if mutation == "foreign_generation" {
				diagnosis["connectionGeneration"] = "other-generation"
			}
			observed.Set("last_error", diagnosis)
			observed.Set("status", "rate_limited")
			if err := f.app.Save(observed); err != nil {
				t.Fatal(err)
			}
			var auth map[string]any
			observed, auth = f.storedAuth(t)
			session := &diagnosisListSession{arrived: make(chan struct{}), resume: make(chan struct{})}
			cleanupConnectionBarrier(t, session.resume)
			completedAt := time.Now().UTC().Truncate(time.Millisecond)
			result := make(chan error, 1)
			go func() {
				_, err := syncPluginCapability(context.Background(), f.app, nil, &pluginSyncRuntimeSession{session: session}, f.plugin, pluginsystem.CapabilityManifest{Name: "list_routes", Export: "list"}, pluginsystem.CapabilityManifest{Name: "get_route_detail", Export: "detail"}, observed, nil, auth, nil, nil, false, false)
				if err == nil {
					err = completePluginSync(f.app, observed, completedAt)
				}
				result <- err
			}()
			waitConnectionBarrier(t, session.arrived)
			flow := startConnectionFlow(t, f, true)
			before, err := pluginsystem.PatchConnection(f.app, c, true, func(fresh *core.Record) error {
				switch mutation {
				case "new_state":
					fresh.Set("state", map[string]any{"cursor": "cursor-sentinel", "nextCapability": "must-survive"})
				case "changed_observed_state":
					fresh.Set("state", map[string]any{"cursor": "advanced-concurrently", "nextCapability": "must-survive"})
				case "changed_error":
					fresh.Set("last_error", map[string]any{"code": "unauthorized", "operation": pluginsystem.OperationList, "authContext": c.AuthContext, "connectionGeneration": c.Generation})
				case "changed_backoff":
					fresh.Set("retry_not_before", time.Now().Add(2*time.Hour))
				case "changed_token":
					fresh.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			before, _ = f.storedAuth(t)
			close(session.resume)
			select {
			case err := <-result:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("completed list did not finish")
			}
			if session.calls.Load() != 1 {
				t.Fatal("completed list was replayed")
			}
			fresh, auth := f.storedAuth(t)
			assertRuntimeRecordChanges(t, before, fresh, "state", "status", "last_error", "retry_not_before", "last_sync_at")
			if auth["oauthState"] != flow["state"] {
				t.Fatal("completion lost pending flow")
			}
			if !fresh.GetDateTime("last_sync_at").Time().Equal(completedAt) {
				t.Fatal("completion did not write its completed time")
			}
			wantState := map[string]any{}
			if mutation == "new_state" {
				wantState = map[string]any{"nextCapability": "must-survive"}
			}
			if mutation == "changed_observed_state" {
				wantState = map[string]any{"cursor": "advanced-concurrently", "nextCapability": "must-survive"}
			}
			if !reflect.DeepEqual(pluginsystem.JSONMapFromRecord(fresh, "state"), wantState) {
				t.Fatal("completion cleared an unobserved/changed cursor, or failed to clear its observed cursor")
			}
			clear := mutation != "changed_error" && mutation != "changed_backoff" && mutation != "changed_token" && mutation != "foreign_context" && mutation != "foreign_generation"
			if clear {
				if fresh.GetString("status") != "configured" || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || !fresh.GetDateTime("retry_not_before").IsZero() {
					t.Fatal("complete unchanged sync failed to clear its observed/historical diagnosis")
				}
			} else {
				for _, field := range []string{"status", "last_error", "retry_not_before"} {
					if !reflect.DeepEqual(before.Get(field), fresh.Get(field)) {
						t.Errorf("completion changed unobserved diagnosis tuple field %s", field)
					}
				}
			}
		})
	}
}

func TestOAuthRefreshClearsOnlyExactObservedRefreshDiagnosis(t *testing.T) {
	for _, mutation := range []string{"exact", "historical", "list_error", "foreign_context", "foreign_generation", "foreign_token", "changed_error", "changed_backoff", "disable"} {
		t.Run(mutation, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			record := seedRuntimeSideFields(t, f)
			c := pluginsystem.CaptureConnection(record)
			oldT := record.GetString("oauth_token_revision")
			diagnosis := map[string]any{"code": "unauthorized", "operation": pluginsystem.OperationOAuthRefresh, "authContext": c.AuthContext, "connectionGeneration": c.Generation, "oauthTokenRevision": oldT}
			switch mutation {
			case "historical":
				diagnosis = map[string]any{"code": "unauthorized", "operation": pluginsystem.OperationOAuthRefresh}
			case "list_error":
				diagnosis["operation"] = pluginsystem.OperationList
			case "foreign_context":
				diagnosis["authContext"] = "other-context"
			case "foreign_generation":
				diagnosis["connectionGeneration"] = "other-generation"
			case "foreign_token":
				diagnosis["oauthTokenRevision"] = "other-token"
			}
			record.Set("last_error", diagnosis)
			record.Set("status", "needs_reauth")
			if err := f.app.Save(record); err != nil {
				t.Fatal(err)
			}
			record, auth := f.storedAuth(t)
			arrived, resume := make(chan struct{}), make(chan struct{})
			cleanupConnectionBarrier(t, resume)
			f.tokenHandler = func(http.ResponseWriter, *http.Request, map[string]string) bool {
				close(arrived)
				<-resume
				return false
			}
			result := make(chan error, 1)
			go func() {
				_, err := pluginsystem.RefreshOAuthToken(context.Background(), f.app, f.plugin, record, auth, c.AuthContext)
				result <- err
			}()
			waitConnectionBarrier(t, arrived)
			flow := startConnectionFlow(t, f, true)
			_, err := pluginsystem.PatchConnection(f.app, c, false, func(fresh *core.Record) error {
				fresh.Set("state", map[string]any{"concurrentCursor": "preserve"})
				fresh.Set("config", map[string]any{"host": map[string]any{"planned": true, "completed": false}})
				switch mutation {
				case "changed_error":
					d := pluginsystem.JSONMapFromRecord(fresh, "last_error")
					d["code"] = "rate_limited"
					fresh.Set("last_error", d)
				case "changed_backoff":
					fresh.Set("retry_not_before", time.Now().Add(2*time.Hour))
				case "disable":
					fresh.Set("enabled", false)
					fresh.Set("status", "disabled")
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			before, _ := f.storedAuth(t)
			close(resume)
			select {
			case err := <-result:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("refresh did not finish")
			}
			fresh, stored := f.storedAuth(t)
			assertRuntimeRecordChanges(t, before, fresh, "auth", "oauth_token_revision", "last_error", "retry_not_before", "status")
			allowed := append([]string{"accessToken", "refreshToken"}, pluginsystem.InternalOAuthMetadataFields()...)
			assertRuntimeAuthChanges(t, before, fresh, allowed...)
			if fresh.GetString("oauth_token_revision") == oldT || stored["accessToken"] != "strava-test-access-rotated" || stored["oauthState"] != flow["state"] || f.tokenRequests.Load() != 1 {
				t.Fatal("refresh lost token/flow, failed to rotate T, or replayed provider")
			}
			clear := mutation == "exact" || mutation == "disable"
			if clear {
				wantStatus := "configured"
				if mutation == "disable" {
					wantStatus = "disabled"
				}
				if fresh.GetString("status") != wantStatus || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || !fresh.GetDateTime("retry_not_before").IsZero() {
					t.Fatal("refresh failed targeted clear or lost concurrent disable")
				}
			} else {
				for _, field := range []string{"status", "last_error", "retry_not_before"} {
					if !reflect.DeepEqual(before.Get(field), fresh.Get(field)) {
						t.Errorf("refresh cleared unrelated/concurrent diagnosis tuple field %s", field)
					}
				}
			}
		})
	}
}
