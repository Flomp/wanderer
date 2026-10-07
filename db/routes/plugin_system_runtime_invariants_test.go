package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

// Compare persisted records after independent reloads. System timestamps are
// omitted; every connection field must be unchanged unless the writer owns it.
func assertRuntimeRecordChanges(t *testing.T, before, after *core.Record, allowed ...string) {
	t.Helper()
	permitted := map[string]bool{"updated": true, "created": true}
	for _, field := range allowed {
		permitted[field] = true
	}
	left, right := before.PublicExport(), after.PublicExport()
	for field, value := range left {
		if !permitted[field] && !reflect.DeepEqual(value, right[field]) {
			t.Errorf("writer changed unowned persisted field %s", field)
		}
	}
	for field := range right {
		if _, present := left[field]; !present && !permitted[field] {
			t.Errorf("writer introduced unowned persisted field %s", field)
		}
	}
}

func assertRuntimeAuthChanges(t *testing.T, before, after *core.Record, allowed ...string) {
	t.Helper()
	permitted := map[string]bool{}
	for _, field := range allowed {
		permitted[field] = true
	}
	left, right := pluginsystem.JSONMapFromRecord(before, "auth"), pluginsystem.JSONMapFromRecord(after, "auth")
	for field, value := range left {
		if !permitted[field] && !reflect.DeepEqual(value, right[field]) {
			t.Errorf("writer changed unowned stored auth field %s", field)
		}
	}
	for field := range right {
		if _, present := left[field]; !present && !permitted[field] {
			t.Errorf("writer introduced unowned stored auth field %s", field)
		}
	}
}

func seedRuntimeSideFields(t *testing.T, f *oauthMetadataFixture) *core.Record {
	t.Helper()
	record, _ := f.storedAuth(t)
	record.Set("config", map[string]any{"host": map[string]any{"planned": false}, "plugin": map[string]any{"note": "settings-sentinel"}})
	record.Set("state", map[string]any{"cursor": "cursor-sentinel"})
	record.Set("last_error", map[string]any{"code": "provider_unavailable", "operation": pluginsystem.OperationList})
	record.Set("retry_not_before", time.Now().Add(time.Hour))
	record.Set("last_sync_at", time.Now().Add(-time.Hour))
	if err := f.app.Save(record); err != nil {
		t.Fatal(err)
	}
	record, _ = f.storedAuth(t)
	return record
}

func TestCallbackActivationIntentCommitsEnabledAndOwnedFields(t *testing.T) {
	for _, tc := range []struct {
		name                            string
		activate, enabled, missingGrant bool
	}{
		{"disabled_grant_without_activation", false, false, false},
		{"new_connection_and_activate", true, false, true},
		{"active_connection_repair", true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			before := seedRuntimeSideFields(t, f)
			before.Set("enabled", tc.enabled)
			before.Set("status", "disabled")
			if tc.enabled {
				before.Set("status", "needs_reauth")
			}
			if tc.missingGrant {
				if err := pluginsystem.ApplyAuthPatch(before, f.plugin.Manifest, map[string]any{"accessToken": nil, "refreshToken": nil, "oauthContext": nil}); err != nil {
					t.Fatal(err)
				}
			}
			if err := f.app.Save(before); err != nil {
				t.Fatal(err)
			}
			before, _ = f.storedAuth(t)
			flow := startConnectionFlow(t, f, tc.activate)
			pending, _ := f.storedAuth(t)
			assertRuntimeRecordChanges(t, before, pending, "auth")
			assertRuntimeAuthChanges(t, before, pending, pluginsystem.InternalOAuthTransientFields()...)
			arrived, resume := make(chan struct{}), make(chan struct{})
			cleanupConnectionBarrier(t, resume)
			f.tokenHandler = func(http.ResponseWriter, *http.Request, map[string]string) bool {
				close(arrived)
				<-resume
				return false
			}
			result := make(chan connectionHTTPResult, 1)
			go func() {
				result <- connectionRequest(f, http.MethodPost, "/api/plugin-system/oauth/callback", map[string]any{"instanceId": f.instanceID, "state": flow["state"], "code": "test-authorization-code"})
			}()
			waitConnectionBarrier(t, arrived)
			// A real Settings write during exchange must survive callback's fresh
			// transaction, including the grant/client side fields it does not own.
			f.request(t, http.MethodPatch, "/api/plugins/instances/"+f.instanceID+"/settings", map[string]any{"config": map[string]any{"host": map[string]any{"completed": false}}})
			concurrent, _ := f.storedAuth(t)
			close(resume)
			response := waitConnectionResult(t, result)
			if response.status != http.StatusOK {
				t.Fatalf("callback status %d", response.status)
			}
			fresh, auth := f.storedAuth(t)
			wantStatus := "disabled"
			if tc.activate {
				wantStatus = "configured"
			}
			if fresh.GetBool("enabled") != (tc.enabled || tc.activate) || fresh.GetString("status") != wantStatus || auth["accessToken"] != "strava-test-access-rotated" {
				t.Fatal("callback did not apply positive/negative activation intent with its grant")
			}
			if fresh.GetString("connection_generation") == before.GetString("connection_generation") || fresh.GetString("oauth_token_revision") == before.GetString("oauth_token_revision") || len(pluginsystem.JSONMapFromRecord(fresh, "state")) != 0 || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || !fresh.GetDateTime("retry_not_before").IsZero() {
				t.Fatal("callback did not reset connection-owned state/revisions")
			}
			assertRuntimeRecordChanges(t, concurrent, fresh, "auth", "enabled", "status", "auth_context", "connection_generation", "oauth_token_revision", "state", "last_error", "retry_not_before")
			allowed := append([]string{"accessToken", "refreshToken"}, pluginsystem.InternalOAuthMetadataFields()...)
			allowed = append(allowed, pluginsystem.InternalOAuthTransientFields()...)
			assertRuntimeAuthChanges(t, concurrent, fresh, allowed...)
			for _, field := range pluginsystem.InternalOAuthTransientFields() {
				if _, present := pluginsystem.JSONMapFromRecord(fresh, "auth")[field]; present {
					t.Errorf("callback retained consumed flow field %s", field)
				}
			}
		})
	}
}

func TestRevokeRequiresOpeningGenerationAndPreservesUnownedFields(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	before := seedRuntimeSideFields(t, f)
	startConnectionFlow(t, f, true)
	before, _ = f.storedAuth(t)
	for _, tc := range []struct {
		name       string
		generation string
		status     int
		code       string
	}{
		{"missing", "", http.StatusBadRequest, "connection_generation_required"},
		{"stale", "obsolete-generation", http.StatusConflict, "plugin_connection_conflict"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := connectionRequest(f, http.MethodPost, "/api/plugin-system/oauth/revoke", map[string]any{"instanceId": f.instanceID, "expected_connection_generation": tc.generation})
			if response.status != tc.status || response.body["data"].(map[string]any)["operation"].(map[string]any)["code"] != tc.code {
				t.Fatalf("revoke guard: %#v", response)
			}
			fresh, _ := f.storedAuth(t)
			assertRuntimeRecordChanges(t, before, fresh)
		})
	}
	release, acquired := pluginsystem.TryLockInstance(f.instanceID)
	if !acquired {
		t.Fatal("test lock unavailable")
	}
	defer release()
	response := connectionRequest(f, http.MethodPost, "/api/plugin-system/oauth/revoke", map[string]any{"instanceId": f.instanceID, "expected_connection_generation": before.GetString("connection_generation")})
	if response.status != http.StatusOK {
		t.Fatalf("revoke status %d", response.status)
	}
	fresh, auth := f.storedAuth(t)
	if fresh.GetBool("enabled") || fresh.GetString("status") != "disabled" || fresh.GetString("connection_generation") == before.GetString("connection_generation") || fresh.GetString("oauth_token_revision") == before.GetString("oauth_token_revision") || response.body["connection_generation"] != fresh.GetString("connection_generation") {
		t.Fatal("revoke did not atomically disable and replace revisions")
	}
	if len(pluginsystem.JSONMapFromRecord(fresh, "state")) != 0 || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 || !fresh.GetDateTime("retry_not_before").IsZero() {
		t.Fatal("revoke retained connection state")
	}
	allowed := append([]string{"accessToken", "refreshToken"}, pluginsystem.InternalOAuthMetadataFields()...)
	allowed = append(allowed, pluginsystem.InternalOAuthTransientFields()...)
	for _, field := range allowed {
		if _, present := auth[field]; present {
			t.Errorf("revoke retained grant/flow field %s", field)
		}
	}
	assertRuntimeAuthChanges(t, before, fresh, allowed...)
	assertRuntimeRecordChanges(t, before, fresh, "auth", "enabled", "status", "connection_generation", "oauth_token_revision", "state", "last_error", "retry_not_before")
	if f.tokenRequests.Load() != 0 {
		t.Fatal("revoke contacted OAuth provider")
	}
}

type cronLockProbeApp struct {
	core.App
	instanceID string
	reads      atomic.Int32
	unlocked   atomic.Bool
}

func (a *cronLockProbeApp) FindRecordById(collection any, id string, filters ...func(*dbx.SelectQuery) error) (*core.Record, error) {
	if collection == "plugin_instances" && id == a.instanceID {
		a.reads.Add(1)
		if !pluginsystem.InstanceBusy(id) {
			a.unlocked.Store(true)
		}
	}
	return a.App.FindRecordById(collection, id, filters...)
}

func TestCronLocksBeforeCapturingFreshDiscoveryConnection(t *testing.T) {
	for _, mutation := range []string{"replace", "disable", "delete"} {
		t.Run(mutation, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			stale, _ := f.storedAuth(t)
			fresh := stale.Fresh()
			switch mutation {
			case "replace":
				pluginsystem.ResetConnection(fresh)
				if err := f.app.Save(fresh); err != nil {
					t.Fatal(err)
				}
			case "disable":
				fresh.Set("enabled", false)
				if err := f.app.Save(fresh); err != nil {
					t.Fatal(err)
				}
			case "delete":
				if err := f.app.Delete(fresh); err != nil {
					t.Fatal(err)
				}
			}
			probe := &cronLockProbeApp{App: f.app, instanceID: f.instanceID}
			_, err := syncPluginInstance(context.Background(), probe, nil, f.plugin, stale)
			if probe.reads.Load() == 0 || probe.unlocked.Load() || pluginsystem.InstanceBusy(f.instanceID) {
				t.Fatal("Cron did not fresh-load under its operation lock, or retained the lock")
			}
			if mutation == "replace" {
				// The intentionally absent actor ends preparation before runtime. An
				// old discovery C would fail first instead of using this new C.
				if err == nil || errors.Is(err, pluginsystem.ErrConnectionObsolete) {
					t.Fatalf("Cron reused old discovery generation: %v", err)
				}
				current, _ := f.storedAuth(t)
				if pluginsystem.JSONMapFromRecord(current, "last_error")["connectionGeneration"] != fresh.GetString("connection_generation") {
					t.Fatal("Cron diagnosis did not use its fresh opening C")
				}
			} else if mutation == "disable" && !errors.Is(err, pluginsystem.ErrInstanceDisabled) {
				t.Fatalf("Cron did not respect latest disable: %v", err)
			} else if mutation == "delete" && !errors.Is(err, pluginsystem.ErrConnectionObsolete) {
				t.Fatalf("Cron did not skip deleted discovery instance: %v", err)
			}
			if f.tokenRequests.Load() != 0 {
				t.Fatal("Cron stale discovery reached provider")
			}
		})
	}
}

type guardedSendRuntime struct{ session *guardedSendSession }
type guardedSendSession struct {
	spec         pluginsystem.HostRequestSpec
	beforeReturn func()
	refreshError error
	calls        atomic.Int32
}

func (*guardedSendRuntime) Call(context.Context, pluginsystem.LocalPlugin, string, []byte, pluginsystem.RequestPolicyContext) ([]byte, error) {
	return nil, errors.New("unexpected direct runtime call")
}
func (r *guardedSendRuntime) OpenSession(context.Context, pluginsystem.LocalPlugin, pluginsystem.RequestPolicyContext) (pluginsystem.RuntimeSession, error) {
	return r.session, nil
}
func (*guardedSendSession) Close(context.Context) error { return nil }
func (s *guardedSendSession) Call(_ context.Context, export string, _ []byte) ([]byte, error) {
	s.calls.Add(1)
	if export == "refresh_auth_v1" {
		return nil, s.refreshError
	}
	if s.beforeReturn != nil {
		s.beforeReturn()
	}
	return json.Marshal(pluginsystem.TrailSendPlan{Request: s.spec})
}

func TestSendRejectsConnectionChangeAfterGuestBeforeUpload(t *testing.T) {
	for _, mutation := range []string{"generation", "context", "disable", "delete"} {
		t.Run(mutation, func(t *testing.T) {
			f := newInteractivePluginFixture(t, nil)
			trail := newInteractiveDiagnosticTrail(t, f)
			arrived, resume := make(chan struct{}), make(chan struct{})
			cleanupConnectionBarrier(t, resume)
			session := &guardedSendSession{spec: pluginsystem.HostRequestSpec{Method: http.MethodPost, Target: pluginsystem.RequestTarget{Type: "connector", Connector: "api", Path: "/upload"}}, beforeReturn: func() { close(arrived); <-resume }}
			var uploads atomic.Int32
			var output bytes.Buffer
			handler := interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{"/api/test/send": func(e *core.RequestEvent) error {
				return pluginSystemTrailSendWithHostRequest(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) { return &guardedSendRuntime{session}, nil }, func(context.Context, pluginsystem.Manifest, pluginsystem.RequestPolicyContext, pluginsystem.HostRequestSpec, pluginsystem.HostRequestOptions) (pluginsystem.HostResponse, error) {
					uploads.Add(1)
					return pluginsystem.HostResponse{Status: 200}, nil
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
			waitConnectionBarrier(t, arrived)
			fresh, err := f.app.FindRecordById("plugin_instances", f.instance.Id)
			if err != nil {
				t.Fatal(err)
			}
			if mutation == "delete" {
				err = f.app.Delete(fresh)
			} else {
				switch mutation {
				case "generation":
					fresh.Set("connection_generation", pluginsystem.NewConnectionRevision())
				case "context":
					fresh.Set("auth_context", "new-context")
				case "disable":
					fresh.Set("enabled", false)
				}
				err = f.app.Save(fresh)
			}
			if err != nil {
				t.Fatal(err)
			}
			before := fresh.Fresh()
			close(resume)
			response := waitConnectionResult(t, result)
			wantCode := "plugin_connection_conflict"
			if mutation == "disable" {
				wantCode = "plugin_instance_disabled"
			}
			if response.status != http.StatusConflict || response.body["data"].(map[string]any)["operation"].(map[string]any)["code"] != wantCode || uploads.Load() != 0 || session.calls.Load() != 1 {
				t.Fatalf("stale Send was not stopped before upload: status=%d uploads=%d calls=%d", response.status, uploads.Load(), session.calls.Load())
			}
			if mutation != "delete" {
				current, err := f.app.FindRecordById("plugin_instances", f.instance.Id)
				if err != nil {
					t.Fatal(err)
				}
				assertRuntimeRecordChanges(t, before, current)
			}
			if pluginsystem.InstanceBusy(f.instance.Id) {
				t.Fatal("Send retained operation lock")
			}
		})
	}
}

func TestSendUploadFailurePreservesConnectionButAuthRefreshFailureDiagnosesIt(t *testing.T) {
	for _, failure := range []string{"upload_http", "upload_network", "auth_refresh"} {
		t.Run(failure, func(t *testing.T) {
			f := newInteractivePluginFixture(t, nil)
			collection := f.instance.Collection()
			collection.Fields.Add(&core.TextField{Name: "status"}, &core.JSONField{Name: "last_error"}, &core.DateField{Name: "retry_not_before"}, &core.JSONField{Name: "state"}, &core.DateField{Name: "last_sync_at"})
			if err := f.app.Save(collection); err != nil {
				t.Fatal(err)
			}
			f.instance.Set("status", "configured")
			f.instance.Set("last_error", map[string]any{"code": "historic-sentinel"})
			f.instance.Set("retry_not_before", time.Now().Add(time.Hour))
			f.instance.Set("state", map[string]any{"cursor": "keep"})
			f.instance.Set("last_sync_at", "")
			if err := f.app.Save(f.instance); err != nil {
				t.Fatal(err)
			}
			before, err := f.app.FindRecordById("plugin_instances", f.instance.Id)
			if err != nil {
				t.Fatal(err)
			}
			trail := newInteractiveDiagnosticTrail(t, f)
			session := &guardedSendSession{spec: pluginsystem.HostRequestSpec{Method: http.MethodPost, Target: pluginsystem.RequestTarget{Type: "connector", Connector: "api", Path: "/upload"}}, refreshError: pluginsystem.Failure("auth_failed", pluginsystem.OperationLogin)}
			if failure == "auth_refresh" {
				session.spec.Auth = "dashboard"
			}
			var uploads atomic.Int32
			var output bytes.Buffer
			handler := interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{"/api/test/send": func(e *core.RequestEvent) error {
				return pluginSystemTrailSendWithHostRequest(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) { return &guardedSendRuntime{session}, nil }, func(context.Context, pluginsystem.Manifest, pluginsystem.RequestPolicyContext, pluginsystem.HostRequestSpec, pluginsystem.HostRequestOptions) (pluginsystem.HostResponse, error) {
					uploads.Add(1)
					if failure == "upload_network" {
						return pluginsystem.HostResponse{}, pluginsystem.WithDiagnosticCause(pluginsystem.Failure("provider_unavailable", pluginsystem.OperationHostHTTP), "network_unavailable")
					}
					return pluginsystem.HostResponse{Status: 429, HeaderValues: http.Header{"Retry-After": []string{"120"}}}, nil
				})
			}})
			token, err := f.owner.NewAuthToken()
			if err != nil {
				t.Fatal(err)
			}
			status, _ := interactiveDiagnosticRequest(t, handler, token, "/api/test/send", map[string]any{"pluginId": f.manifest.ID, "trailId": trail.Id})
			if status != http.StatusBadRequest {
				t.Fatalf("failure status %d", status)
			}
			fresh, err := f.app.FindRecordById("plugin_instances", f.instance.Id)
			if err != nil {
				t.Fatal(err)
			}
			if failure == "auth_refresh" {
				if uploads.Load() != 0 || fresh.GetString("status") != "needs_reauth" || pluginsystem.JSONMapFromRecord(fresh, "last_error")["operation"] != pluginsystem.OperationLogin {
					t.Fatal("auth refresh failure did not retain its connection diagnosis")
				}
				assertRuntimeRecordChanges(t, before, fresh, "status", "last_error", "retry_not_before")
			} else {
				if uploads.Load() != 1 {
					t.Fatal("upload was not executed exactly once")
				}
				assertRuntimeRecordChanges(t, before, fresh)
			}
		})
	}
}
