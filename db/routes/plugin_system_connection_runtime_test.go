package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

type connectionHTTPResult struct {
	status int
	body   map[string]any
}

func connectionRequest(f *oauthMetadataFixture, method, path string, body map[string]any) connectionHTTPResult {
	data, _ := json.Marshal(body)
	request := httptest.NewRequest(method, path, bytes.NewReader(data))
	request.Header.Set("Authorization", f.authToken)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, request)
	result := connectionHTTPResult{status: response.Code}
	_ = json.Unmarshal(response.Body.Bytes(), &result.body)
	return result
}

func waitConnectionBarrier(t *testing.T, arrived <-chan struct{}) {
	t.Helper()
	select {
	case <-arrived:
	case <-time.After(10 * time.Second):
		t.Fatal("operation did not reach provider barrier")
	}
}

func cleanupConnectionBarrier(t *testing.T, resume chan struct{}) {
	t.Helper()
	t.Cleanup(func() {
		select {
		case <-resume:
		default:
			close(resume)
		}
	})
}

func waitConnectionResult(t *testing.T, result <-chan connectionHTTPResult) connectionHTTPResult {
	t.Helper()
	select {
	case value := <-result:
		return value
	case <-time.After(10 * time.Second):
		t.Fatal("operation failed to finish")
		return connectionHTTPResult{}
	}
}

func startConnectionFlow(t *testing.T, f *oauthMetadataFixture, activate bool) map[string]any {
	t.Helper()
	return f.request(t, http.MethodPost, "/api/plugin-system/oauth/start", map[string]any{"pluginId": "strava", "instanceId": f.instanceID, "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback", "activate_on_success": activate})
}

func TestOAuthStartKeepsActiveSelectionAndGrantUntilAtomicCallback(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	before, beforeAuth := f.storedAuth(t)
	manifest := f.plugin.Manifest
	manifest.Auth.Contexts["other"] = manifest.Auth.Contexts["oauth_access_token"]
	manifest.Permissions.Auth = append(manifest.Permissions.Auth, "other")
	installed, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", "strava")
	if err != nil {
		t.Fatal(err)
	}
	installed.Set("manifest", manifest)
	if err := f.app.Save(installed); err != nil {
		t.Fatal(err)
	}
	flow := f.request(t, http.MethodPost, "/api/plugin-system/oauth/start", map[string]any{"pluginId": "strava", "instanceId": f.instanceID, "authContext": "other", "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback"})
	pending, pendingAuth := f.storedAuth(t)
	for _, field := range []string{"auth_context", "connection_generation", "oauth_token_revision", "status"} {
		if pending.GetString(field) != before.GetString(field) {
			t.Fatalf("start changed %s", field)
		}
	}
	for _, field := range []string{"accessToken", "refreshToken", "oauthContext"} {
		if pendingAuth[field] != beforeAuth[field] {
			t.Fatalf("start changed grant %s", field)
		}
	}
	if pendingAuth["oauthFlowContext"] != "other" {
		t.Fatal("flow proposal not recorded")
	}
	f.request(t, http.MethodPost, "/api/plugin-system/oauth/callback", map[string]any{"instanceId": f.instanceID, "state": flow["state"], "code": "test-authorization-code"})
	committed, committedAuth := f.storedAuth(t)
	if committed.GetString("auth_context") != "other" || committedAuth["oauthContext"] != "other" || committed.GetString("connection_generation") == before.GetString("connection_generation") || committed.GetString("oauth_token_revision") == before.GetString("oauth_token_revision") {
		t.Fatal("callback did not atomically replace selection/grant/revisions")
	}
}

func TestCallbackUsesFreshActivationIntentAndIgnoresProviderLock(t *testing.T) {
	for _, patches := range [][]bool{{false}, {false, false}, {false, true, false}} {
		t.Run(string(rune('0'+len(patches))), func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			flow := startConnectionFlow(t, f, true)
			arrived, resume := make(chan struct{}), make(chan struct{})
			cleanupConnectionBarrier(t, resume)
			f.tokenHandler = func(http.ResponseWriter, *http.Request, map[string]string) bool {
				close(arrived)
				<-resume
				return false
			}
			release, acquired := pluginsystem.TryLockInstance(f.instanceID)
			if !acquired {
				t.Fatal("test lock unavailable")
			}
			defer release()
			result := make(chan connectionHTTPResult, 1)
			go func() {
				result <- connectionRequest(f, http.MethodPost, "/api/plugin-system/oauth/callback", map[string]any{"instanceId": f.instanceID, "state": flow["state"], "code": "test-authorization-code"})
			}()
			waitConnectionBarrier(t, arrived)
			for _, enabled := range patches {
				f.request(t, http.MethodPatch, "/api/plugins/instances/"+f.instanceID+"/settings", map[string]any{"enabled": enabled})
			}
			close(resume)
			if response := waitConnectionResult(t, result); response.status != 200 {
				t.Fatalf("callback: %#v", response)
			}
			fresh, auth := f.storedAuth(t)
			if fresh.GetBool("enabled") || fresh.GetString("status") != "disabled" || auth["accessToken"] != "strava-test-access-rotated" {
				t.Fatal("callback lost disable or successful grant")
			}
		})
	}
}

func TestConcurrentSuccessfulAndFailedCallbackKeepSuccessfulGrant(t *testing.T) {
	for _, failureFirst := range []bool{true, false} {
		name := "success_first"
		if failureFirst {
			name = "failure_first"
		}
		t.Run(name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			flow := startConnectionFlow(t, f, false)
			failedArrived, successArrived := make(chan struct{}), make(chan struct{})
			failedResume, successResume := make(chan struct{}), make(chan struct{})
			cleanupConnectionBarrier(t, failedResume)
			cleanupConnectionBarrier(t, successResume)
			f.tokenHandler = func(w http.ResponseWriter, _ *http.Request, body map[string]string) bool {
				if body["code"] == "failed" {
					close(failedArrived)
					<-failedResume
					w.WriteHeader(400)
					_, _ = w.Write([]byte(`{"error":"invalid_grant","private":"provider-secret"}`))
					return true
				}
				close(successArrived)
				<-successResume
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"access_token":"accepted-grant","refresh_token":"accepted-refresh"}`))
				return true
			}
			failed, success := make(chan connectionHTTPResult, 1), make(chan connectionHTTPResult, 1)
			go func() {
				failed <- connectionRequest(f, http.MethodPost, "/api/plugin-system/oauth/callback", map[string]any{"instanceId": f.instanceID, "state": flow["state"], "code": "failed"})
			}()
			go func() {
				success <- connectionRequest(f, http.MethodPost, "/api/plugin-system/oauth/callback", map[string]any{"instanceId": f.instanceID, "state": flow["state"], "code": "success"})
			}()
			waitConnectionBarrier(t, failedArrived)
			waitConnectionBarrier(t, successArrived)
			if failureFirst {
				close(failedResume)
				if response := waitConnectionResult(t, failed); response.status == 200 {
					t.Fatal("failed exchange succeeded")
				}
				close(successResume)
				if response := waitConnectionResult(t, success); response.status != 200 {
					t.Fatalf("success: %#v", response)
				}
			} else {
				close(successResume)
				if response := waitConnectionResult(t, success); response.status != 200 {
					t.Fatalf("success: %#v", response)
				}
				close(failedResume)
				if response := waitConnectionResult(t, failed); response.status == 200 {
					t.Fatal("failed exchange succeeded")
				}
			}
			fresh, auth := f.storedAuth(t)
			if auth["accessToken"] != "accepted-grant" || len(pluginsystem.JSONMapFromRecord(fresh, "last_error")) != 0 {
				t.Fatal("failed callback lost grant or changed diagnosis")
			}
			f.assertAuthFieldsAbsent(t, "consumed flow", "oauthState", "oauthFlowContext", "oauthActivateOnCallback")
		})
	}
}

func TestRefreshPatchesTokensPreservingConcurrentControlWrites(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	instance, auth := f.storedAuth(t)
	snapshot := pluginsystem.CaptureConnection(instance)
	oldT := instance.GetString("oauth_token_revision")
	arrived, resume := make(chan struct{}), make(chan struct{})
	cleanupConnectionBarrier(t, resume)
	f.tokenHandler = func(http.ResponseWriter, *http.Request, map[string]string) bool {
		close(arrived)
		<-resume
		return false
	}
	result := make(chan error, 1)
	go func() {
		_, err := pluginsystem.RefreshOAuthToken(context.Background(), f.app, f.plugin, instance, auth, snapshot.AuthContext)
		result <- err
	}()
	waitConnectionBarrier(t, arrived)
	flow := startConnectionFlow(t, f, true)
	f.request(t, http.MethodPatch, "/api/plugins/instances/"+f.instanceID+"/settings", map[string]any{"enabled": false, "config": map[string]any{"host": map[string]any{"planned": false}}})
	close(resume)
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("refresh stuck")
	}
	fresh, stored := f.storedAuth(t)
	if fresh.GetBool("enabled") || fresh.GetString("status") != "disabled" || fresh.GetString("connection_generation") != snapshot.Generation || fresh.GetString("oauth_token_revision") == oldT {
		t.Fatal("refresh changed current control state or failed to rotate T")
	}
	if stored["oauthState"] != flow["state"] || stored["oauthActivateOnCallback"] != false || stored["accessToken"] != "strava-test-access-rotated" {
		t.Fatal("refresh lost newest flow or grant")
	}
	if !reflect.DeepEqual(pluginsystem.JSONMapFromRecord(fresh, "config"), map[string]any{"host": map[string]any{"planned": false}}) {
		t.Fatal("refresh overwrote config")
	}
}

func TestRefreshTokenConflictReloadsWithoutProviderReplay(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	instance, auth := f.storedAuth(t)
	snapshot := pluginsystem.CaptureConnection(instance)
	arrived, resume := make(chan struct{}), make(chan struct{})
	cleanupConnectionBarrier(t, resume)
	f.tokenHandler = func(http.ResponseWriter, *http.Request, map[string]string) bool {
		close(arrived)
		<-resume
		return false
	}
	type refreshResult struct {
		auth map[string]any
		err  error
	}
	result := make(chan refreshResult, 1)
	go func() {
		auth, err := pluginsystem.RefreshOAuthToken(context.Background(), f.app, f.plugin, instance, auth, snapshot.AuthContext)
		result <- refreshResult{auth, err}
	}()
	waitConnectionBarrier(t, arrived)
	_, err := pluginsystem.PatchConnection(f.app, snapshot, false, func(fresh *core.Record) error {
		if err := pluginsystem.ApplyAuthPatch(fresh, f.plugin.Manifest, map[string]any{"accessToken": "newer-token", "refreshToken": "newer-refresh"}); err != nil {
			return err
		}
		fresh.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	close(resume)
	select {
	case response := <-result:
		if response.err != nil || response.auth["accessToken"] != "newer-token" {
			t.Fatalf("stale refresh result: %#v", response)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("refresh stuck")
	}
	if f.tokenRequests.Load() != 1 {
		t.Fatal("token conflict replayed provider refresh")
	}
	_, stored := f.storedAuth(t)
	if stored["accessToken"] != "newer-token" {
		t.Fatal("stale refresh overwrote current tokens")
	}
}

func TestCallbackNewFlowConflictsWithoutGrantWrite(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	_, before := f.storedAuth(t)
	flow := startConnectionFlow(t, f, false)
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
	newFlow := startConnectionFlow(t, f, false)
	close(resume)
	response := waitConnectionResult(t, result)
	if response.status != 409 || response.body["data"].(map[string]any)["operation"].(map[string]any)["code"] != "plugin_oauth_flow_conflict" {
		t.Fatalf("flow conflict: %#v", response)
	}
	_, auth := f.storedAuth(t)
	if auth["accessToken"] != before["accessToken"] || auth["oauthState"] != newFlow["state"] {
		t.Fatal("stale callback changed grant/new flow")
	}
}

func TestOAuthProvenanceMismatchStopsBeforeProviderIO(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	instance, _ := f.storedAuth(t)
	if err := pluginsystem.ApplyAuthPatch(instance, f.plugin.Manifest, map[string]any{"oauthContext": "different-context"}); err != nil {
		t.Fatal(err)
	}
	if err := f.app.Save(instance); err != nil {
		t.Fatal(err)
	}
	if _, err := decryptedActiveInstanceAuth(f.plugin, instance); err == nil {
		t.Fatal("grant mismatch accepted")
	}
	auth := pluginsystem.JSONMapFromRecord(instance, "auth")
	if _, err := pluginsystem.RefreshOAuthAuthIfNeeded(context.Background(), f.app, f.plugin, instance, auth); err == nil {
		t.Fatal("mismatched grant refreshed")
	}
	if f.tokenRequests.Load() != 0 {
		t.Fatal("mismatched grant reached provider")
	}
}

func TestBusySyncPerformsNoWrite(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	instance, _ := f.storedAuth(t)
	before := instance.PublicExport()
	release, acquired := pluginsystem.TryLockInstance(instance.Id)
	if !acquired {
		t.Fatal("unable to hold lock")
	}
	defer release()
	_, err := syncPluginInstance(context.Background(), f.app, nil, f.plugin, instance)
	if !errors.Is(err, pluginsystem.ErrInstanceBusy) {
		t.Fatalf("busy sync: %v", err)
	}
	fresh, _ := f.storedAuth(t)
	if !reflect.DeepEqual(before, fresh.PublicExport()) || f.tokenRequests.Load() != 0 {
		t.Fatal("busy sync changed instance or called provider")
	}
}

func TestRefreshRejectsProviderAccountMismatchWithoutTokenWrite(t *testing.T) {
	f := newOAuthMetadataFixture(t, func(auth map[string]any) { auth["providerUserId"] = "known-account" })
	instance, auth := f.storedAuth(t)
	before := instance.PublicExport()
	f.tokenHandler = func(w http.ResponseWriter, _ *http.Request, body map[string]string) bool {
		if body["grant_type"] != "refresh_token" || body["refresh_token"] != f.expectedRefreshToken {
			t.Error("account mismatch test did not use the real refresh boundary")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"different-account-access","refresh_token":"different-account-refresh","providerUserId":"different-account","expires_in":3600}`))
		return true
	}
	_, err := pluginsystem.RefreshOAuthToken(context.Background(), f.app, f.plugin, instance, auth, instance.GetString("auth_context"))
	if err == nil || pluginsystem.PublicPluginError(err).Code != "invalid_grant" {
		t.Fatalf("account mismatch accepted: %v", err)
	}
	fresh, _ := f.storedAuth(t)
	if !reflect.DeepEqual(before, fresh.PublicExport()) || f.tokenRequests.Load() != 1 {
		t.Fatal("account mismatch changed grant/control state or replayed provider work")
	}
}

func TestRefreshPreservesConcurrentFirstAccountPinning(t *testing.T) {
	for _, tc := range []struct {
		name, returnedID string
		accepted         bool
	}{{"matching_identity", "learned-account", true}, {"absent_identity", "", true}, {"conflicting_identity", "another-account", false}} {
		t.Run(tc.name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			instance, auth := f.storedAuth(t)
			c := pluginsystem.CaptureConnection(instance)
			oldT := instance.GetString("oauth_token_revision")
			arrived, resume := make(chan struct{}), make(chan struct{})
			cleanupConnectionBarrier(t, resume)
			f.tokenHandler = func(w http.ResponseWriter, _ *http.Request, body map[string]string) bool {
				if body["grant_type"] != "refresh_token" {
					t.Error("account pinning test bypassed refresh")
				}
				close(arrived)
				<-resume
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "pinned-account-access", "refresh_token": "pinned-account-refresh", "providerUserId": tc.returnedID, "expires_in": 3600})
				return true
			}
			result := make(chan error, 1)
			go func() {
				_, err := pluginsystem.RefreshOAuthToken(context.Background(), f.app, f.plugin, instance, auth, c.AuthContext)
				result <- err
			}()
			waitConnectionBarrier(t, arrived)
			// Another host operation learns the account from this unchanged token
			// chain. It keeps G/T and gives the final refresh CAS a known pin.
			_, err := pluginsystem.PatchConnection(f.app, c, false, func(fresh *core.Record) error {
				return pluginsystem.ApplyAuthPatch(fresh, f.plugin.Manifest, map[string]any{"providerUserId": "learned-account"})
			})
			if err != nil {
				t.Fatal(err)
			}
			// Compare two database reads, including updated. The in-memory
			// Autodate value can carry a different time location from its stored
			// representation even when they denote exactly the same timestamp.
			pinned, err := f.app.FindRecordById("plugin_instances", f.instanceID)
			if err != nil {
				t.Fatal(err)
			}
			before := pinned.PublicExport()
			close(resume)
			select {
			case err := <-result:
				if tc.accepted && err != nil {
					t.Fatal(err)
				}
				if !tc.accepted && (err == nil || pluginsystem.PublicPluginError(err).Code != "invalid_grant") {
					t.Fatalf("conflicting account accepted: %v", err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("refresh stuck")
			}
			fresh, stored := f.storedAuth(t)
			if stored["providerUserId"] != "learned-account" || fresh.GetString("connection_generation") != c.Generation || f.tokenRequests.Load() != 1 {
				t.Fatal("refresh lost first account pin, changed G, or replayed provider work")
			}
			if tc.accepted {
				if stored["accessToken"] != "pinned-account-access" || fresh.GetString("oauth_token_revision") == oldT {
					t.Fatal("matching token chain did not rotate")
				}
			} else if !reflect.DeepEqual(before, fresh.PublicExport()) {
				t.Fatal("conflicting returned identity overwrote the first account pin")
			}
		})
	}
}

func TestOAuthStartReturnsItsOwnCommittedRevisionAfterConcurrentConnectionChange(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	before, _ := f.storedAuth(t)
	arrived, resume := make(chan struct{}), make(chan struct{})
	cleanupConnectionBarrier(t, resume)
	var paused atomic.Bool
	var startUpdated string
	f.app.OnRecordAfterUpdateSuccess("plugin_instances").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.Id == f.instanceID && paused.CompareAndSwap(false, true) {
			startUpdated = e.Record.GetString("updated")
			close(arrived)
			<-resume
		}
		return e.Next()
	})
	result := make(chan connectionHTTPResult, 1)
	go func() {
		result <- connectionRequest(f, http.MethodPost, "/api/plugin-system/oauth/start", map[string]any{"pluginId": "strava", "instanceId": f.instanceID, "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback", "expected_connection_generation": before.GetString("connection_generation"), "activate_on_success": true})
	}()
	waitConnectionBarrier(t, arrived)
	f.request(t, http.MethodPatch, "/api/plugins/instances/"+f.instanceID+"/settings", map[string]any{"auth": map[string]any{"clientSecret": "concurrent-client-secret"}, "expected_connection_generation": before.GetString("connection_generation")})
	newer, _ := f.storedAuth(t)
	if newer.GetString("connection_generation") == before.GetString("connection_generation") {
		t.Fatal("competing connection writer did not commit")
	}
	close(resume)
	response := waitConnectionResult(t, result)
	if response.status != http.StatusOK || response.body["connection_generation"] != before.GetString("connection_generation") || response.body["updated"] != startUpdated || startUpdated == "" {
		t.Fatalf("start rebound its response to a newer connection: %#v", response)
	}
	if len(response.body) != 5 || response.body["instanceId"] != f.instanceID || response.body["state"] == "" || response.body["url"] == "" {
		t.Fatalf("unexpected start response: %#v", response.body)
	}
}

func TestRevokeReturnsItsOwnCommittedGenerationAfterConcurrentConnectionChange(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	before, _ := f.storedAuth(t)
	arrived, resume := make(chan struct{}), make(chan struct{})
	cleanupConnectionBarrier(t, resume)
	var paused atomic.Bool
	var revokedGeneration string
	var revokedUpdated string
	// PocketBase invokes this success hook after the transaction has committed.
	// Pausing here leaves the DB free for a real competing Settings writer.
	f.app.OnRecordAfterUpdateSuccess("plugin_instances").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.Id == f.instanceID && paused.CompareAndSwap(false, true) {
			revokedGeneration = e.Record.GetString("connection_generation")
			revokedUpdated = e.Record.GetString("updated")
			close(arrived)
			<-resume
		}
		return e.Next()
	})
	result := make(chan connectionHTTPResult, 1)
	go func() {
		result <- connectionRequest(f, http.MethodPost, "/api/plugin-system/oauth/revoke", map[string]any{"instanceId": f.instanceID, "expected_connection_generation": before.GetString("connection_generation")})
	}()
	waitConnectionBarrier(t, arrived)
	if revokedGeneration == "" || revokedGeneration == before.GetString("connection_generation") {
		t.Fatal("revoke did not mint a new generation")
	}
	f.request(t, http.MethodPatch, "/api/plugins/instances/"+f.instanceID+"/settings", map[string]any{"auth": map[string]any{"clientSecret": "concurrent-client-secret"}, "expected_connection_generation": revokedGeneration})
	newer, _ := f.storedAuth(t)
	if newer.GetString("connection_generation") == revokedGeneration {
		t.Fatal("competing connection writer did not commit")
	}
	close(resume)
	response := waitConnectionResult(t, result)
	if response.status != http.StatusOK || response.body["connection_generation"] != revokedGeneration {
		t.Fatalf("revoke rebound its response to the newer connection: %#v", response)
	}
	if len(response.body) != 3 || response.body["ok"] != true || response.body["updated"] != revokedUpdated || revokedUpdated == "" {
		t.Fatalf("revoke exposed unexpected fields: %#v", response.body)
	}
	start := connectionRequest(f, http.MethodPost, "/api/plugin-system/oauth/start", map[string]any{"pluginId": "strava", "instanceId": f.instanceID, "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback", "expected_connection_generation": response.body["connection_generation"]})
	if start.status != http.StatusConflict {
		t.Fatalf("repair Start adopted the competing connection: %#v", start)
	}
	latest, _ := f.storedAuth(t)
	if latest.GetString("connection_generation") != newer.GetString("connection_generation") || f.tokenRequests.Load() != 0 {
		t.Fatal("stale repair Start changed the competing connection or reached the provider")
	}
}
