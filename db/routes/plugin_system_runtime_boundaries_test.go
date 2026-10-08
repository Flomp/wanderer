package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

// A read can become stale immediately after its query. Hold that precise
// boundary rather than mutating during guest execution, where earlier guards
// would mask a missing final pre-upload check.
type runtimeConnectionReadBarrierApp struct {
	core.App
	instanceID    string
	beforeRead    bool
	armed         atomic.Bool
	arrived       chan struct{}
	resume        chan struct{}
	arrivedLocked atomic.Bool
}

func (a *runtimeConnectionReadBarrierApp) pause() {
	a.arrivedLocked.Store(pluginsystem.InstanceBusy(a.instanceID))
	close(a.arrived)
	<-a.resume
}

func (a *runtimeConnectionReadBarrierApp) FindRecordById(collection any, id string, filters ...func(*dbx.SelectQuery) error) (*core.Record, error) {
	if collection != "plugin_instances" || id != a.instanceID || !a.armed.CompareAndSwap(true, false) {
		return a.App.FindRecordById(collection, id, filters...)
	}
	if a.beforeRead {
		a.pause()
	}
	record, err := a.App.FindRecordById(collection, id, filters...)
	if !a.beforeRead {
		a.pause()
	}
	return record, err
}

func TestCronRechecksBackoffFromReloadUnderLock(t *testing.T) {
	for _, mutation := range []string{"set_future_backoff", "clear_discovered_backoff"} {
		t.Run(mutation, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			discovered := seedRuntimeSideFields(t, f)
			if mutation == "set_future_backoff" {
				discovered.Set("retry_not_before", "")
			}
			// A valid actor and connector allow real OAuth provider work if the
			// backoff guard is removed. The cleared-backoff control instead gets
			// an explicit local config rejection after passing that guard.
			actors := core.NewBaseCollection("activitypub_actors")
			actors.Fields.Add(&core.TextField{Name: "user"})
			if err := f.app.Save(actors); err != nil {
				t.Fatal(err)
			}
			actor := core.NewRecord(actors)
			actor.Set("user", discovered.GetString("user"))
			if err := f.app.Save(actor); err != nil {
				t.Fatal(err)
			}
			for i := range f.plugin.Manifest.Permissions.Network.Connectors {
				connector := &f.plugin.Manifest.Permissions.Network.Connectors[i]
				if connector.Name == "api" {
					connector.Type = pluginsystem.ConnectorTypeConfigured
					connector.ConfigKey = "server"
					connector.FixedBaseURL = ""
				}
			}
			discovered.Set("config", map[string]any{"host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": "https://provider.example.test"}}}})
			if err := f.app.Save(discovered); err != nil {
				t.Fatal(err)
			}
			discovered, auth := f.storedAuth(t)
			if !pluginsystem.OAuthNeedsRefresh(auth) {
				t.Fatal("fixture must require provider refresh when backoff permits a run")
			}
			probe := &runtimeConnectionReadBarrierApp{App: f.app, instanceID: f.instanceID, beforeRead: true, arrived: make(chan struct{}), resume: make(chan struct{})}
			probe.armed.Store(true)
			cleanupConnectionBarrier(t, probe.resume)
			result := make(chan error, 1)
			go func() {
				_, err := syncPluginInstance(context.Background(), probe, nil, f.plugin, discovered)
				result <- err
			}()
			waitConnectionBarrier(t, probe.arrived)
			fresh, _ := f.storedAuth(t)
			if mutation == "set_future_backoff" {
				fresh.Set("retry_not_before", time.Now().Add(time.Hour))
				fresh.Set("status", "rate_limited")
			} else {
				fresh.Set("retry_not_before", "")
				fresh.Set("config", map[string]any{"host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": "file:///invalid-config"}}}})
			}
			if err := f.app.Save(fresh); err != nil {
				t.Fatal(err)
			}
			before, _ := f.storedAuth(t)
			close(probe.resume)
			var err error
			select {
			case err = <-result:
			case <-time.After(10 * time.Second):
				t.Fatal("Cron did not finish after releasing reload barrier")
			}
			if !probe.arrivedLocked.Load() || pluginsystem.InstanceBusy(f.instanceID) || f.tokenRequests.Load() != 0 {
				t.Errorf("Cron boundary lock/provider mismatch: locked=%t busy=%t provider=%d", probe.arrivedLocked.Load(), pluginsystem.InstanceBusy(f.instanceID), f.tokenRequests.Load())
			}
			current, _ := f.storedAuth(t)
			if mutation == "set_future_backoff" {
				if !errors.Is(err, pluginsystem.ErrConnectionObsolete) {
					t.Errorf("fresh future backoff did not skip Cron: %v", err)
				}
				assertRuntimeRecordChanges(t, before, current)
			} else {
				if err == nil || pluginsystem.PublicPluginError(err).Code != "invalid_configuration" {
					t.Errorf("Cron used old discovery backoff instead of passing to local config validation: %v", err)
				}
				if current.GetString("status") != "error" || pluginsystem.JSONMapFromRecord(current, "last_error")["code"] != "invalid_configuration" {
					t.Error("cleared-backoff control did not reach local config diagnosis")
				}
				assertRuntimeRecordChanges(t, before, current, "status", "last_error", "retry_not_before")
			}
		})
	}
}

type uploadBoundarySession struct {
	probe        *runtimeConnectionReadBarrierApp
	calls        atomic.Int32
	refreshCalls atomic.Int32
	closes       atomic.Int32
}

func (s *uploadBoundarySession) Call(_ context.Context, export string, _ []byte) ([]byte, error) {
	s.calls.Add(1)
	if export == "refresh_auth_v1" {
		s.refreshCalls.Add(1)
		// The next read is the post-refresh injection guard. Freeze its DB
		// result after the query so injection can finish on that snapshot.
		s.probe.armed.Store(true)
		return []byte(`{"token":"session-result-token","scheme":"Bearer"}`), nil
	}
	return json.Marshal(pluginsystem.TrailSendPlan{Request: pluginsystem.HostRequestSpec{Auth: "dashboard", Method: http.MethodPost, Target: pluginsystem.RequestTarget{Type: "connector", Connector: "api", Path: "/upload"}}})
}

func (s *uploadBoundarySession) Close(context.Context) error {
	s.closes.Add(1)
	return nil
}

type uploadBoundaryRuntime struct{ session *uploadBoundarySession }

func (*uploadBoundaryRuntime) Call(context.Context, pluginsystem.LocalPlugin, string, []byte, pluginsystem.RequestPolicyContext) ([]byte, error) {
	return nil, errors.New("unexpected direct runtime call")
}

func (r *uploadBoundaryRuntime) OpenSession(context.Context, pluginsystem.LocalPlugin, pluginsystem.RequestPolicyContext) (pluginsystem.RuntimeSession, error) {
	return r.session, nil
}

func TestSendFinalConnectionGuardAfterSuccessfulAuthBeforeUpload(t *testing.T) {
	for _, mutation := range []string{"unchanged", "generation", "context", "disable", "delete"} {
		t.Run(mutation, func(t *testing.T) {
			f := newInteractivePluginFixture(t, nil)
			trail := newInteractiveDiagnosticTrail(t, f)
			probe := &runtimeConnectionReadBarrierApp{App: f.app, instanceID: f.instance.Id, arrived: make(chan struct{}), resume: make(chan struct{})}
			cleanupConnectionBarrier(t, probe.resume)
			session := &uploadBoundarySession{probe: probe}
			var uploads atomic.Int32
			var output bytes.Buffer
			handler := interactiveDiagnosticHandler(t, probe, &output, map[string]func(*core.RequestEvent) error{"/api/test/send": func(e *core.RequestEvent) error {
				return pluginSystemTrailSendWithHostRequest(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) {
					return &uploadBoundaryRuntime{session}, nil
				}, func(_ context.Context, _ pluginsystem.Manifest, _ pluginsystem.RequestPolicyContext, spec pluginsystem.HostRequestSpec, opts pluginsystem.HostRequestOptions) (pluginsystem.HostResponse, error) {
					uploads.Add(1)
					if spec.Auth != "" || spec.Headers["Authorization"] != "Bearer session-result-token" || len(opts.Trail) == 0 {
						t.Error("upload control did not receive successfully resolved auth and trail")
					}
					return pluginsystem.HostResponse{Status: http.StatusOK}, nil
				})
			}})
			token, err := f.owner.NewAuthToken()
			if err != nil {
				t.Fatal(err)
			}
			result := make(chan connectionHTTPResult, 1)
			go func() {
				status, body := interactiveDiagnosticRequest(t, handler, token, "/api/test/send", map[string]any{"pluginId": f.manifest.ID, "trailId": trail.Id})
				result <- connectionHTTPResult{status, body}
			}()
			waitConnectionBarrier(t, probe.arrived)
			fresh, err := f.app.FindRecordById("plugin_instances", f.instance.Id)
			if err != nil {
				t.Fatal(err)
			}
			if mutation == "delete" {
				err = f.app.Delete(fresh)
			} else if mutation != "unchanged" {
				switch mutation {
				case "generation":
					fresh.Set("connection_generation", pluginsystem.NewConnectionRevision())
				case "context":
					fresh.Set("auth_context", "replacement-context")
				case "disable":
					fresh.Set("enabled", false)
				}
				err = f.app.Save(fresh)
			}
			if err != nil {
				t.Fatal(err)
			}
			var before *core.Record
			if mutation != "delete" {
				before, err = f.app.FindRecordById("plugin_instances", f.instance.Id)
				if err != nil {
					t.Fatal(err)
				}
			}
			close(probe.resume)
			response := waitConnectionResult(t, result)
			if !probe.arrivedLocked.Load() || pluginsystem.InstanceBusy(f.instance.Id) || session.calls.Load() != 2 || session.refreshCalls.Load() != 1 || session.closes.Load() != 1 {
				t.Error("Send did not complete successful auth exactly once under/releasing its operation lock")
			}
			if mutation == "unchanged" {
				if response.status != http.StatusOK || uploads.Load() != 1 {
					t.Errorf("unchanged authenticated Send did not upload once: status=%d uploads=%d", response.status, uploads.Load())
				}
			} else {
				wantCode := "plugin_connection_conflict"
				if mutation == "disable" {
					wantCode = "plugin_instance_disabled"
				}
				if response.status != http.StatusConflict || response.body["data"].(map[string]any)["operation"].(map[string]any)["code"] != wantCode || uploads.Load() != 0 {
					t.Errorf("final Send guard failed: status=%d uploads=%d", response.status, uploads.Load())
				}
			}
			if mutation == "delete" {
				if current, _ := f.app.FindRecordById("plugin_instances", f.instance.Id); current != nil {
					t.Error("Send recreated deleted instance")
				}
			} else {
				current, err := f.app.FindRecordById("plugin_instances", f.instance.Id)
				if err != nil {
					t.Fatal(err)
				}
				assertRuntimeRecordChanges(t, before, current)
			}
		})
	}
}
