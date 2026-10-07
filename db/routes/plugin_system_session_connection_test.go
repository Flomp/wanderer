package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"pocketbase/pluginsystem"
)

func TestExistingValidateBusyAndExpectedGenerationPerformNoProviderIO(t *testing.T) {
	f := newSessionValidateFixture(t)
	release, acquired := pluginsystem.TryLockInstance(f.instance.Id)
	if !acquired {
		t.Fatal("test lock unavailable")
	}
	status, body := f.request(t, f.body())
	release()
	if status != 409 || !strings.Contains(body, `"code":"plugin_instance_busy"`) || f.runtime.calls.Load() != 0 || f.providerRequests.Load() != 0 {
		t.Fatalf("busy Validate: %d %s", status, body)
	}
	request := f.body()
	delete(request, "expected_connection_generation")
	status, body = f.request(t, request)
	if status != 400 || !strings.Contains(body, `"code":"connection_generation_required"`) {
		t.Fatalf("missing guard: %d %s", status, body)
	}
	request["expected_connection_generation"] = "stale-generation"
	status, body = f.request(t, request)
	if status != 409 || !strings.Contains(body, `"code":"plugin_connection_conflict"`) || f.runtime.calls.Load() != 0 {
		t.Fatalf("stale guard: %d %s", status, body)
	}
}

func TestValidateCredentialPatchNeverValidatesRemovedOrClearedCredentials(t *testing.T) {
	for _, tc := range []struct {
		name    string
		patch   map[string]any
		message string
	}{
		{"null_secret_removes", map[string]any{"password": nil}, "Required session credentials are missing"},
		{"empty_nonsecret_clears", map[string]any{"email": "", "password": "replacement-password"}, "Required session credentials are missing"},
		{"null_nonsecret_rejected", map[string]any{"email": nil, "password": "replacement-password"}, "Only editable secret credentials"},
		{"host_metadata_rejected", map[string]any{"oauthState": "injected-state", "password": "replacement-password"}, "Unsupported credential field"},
		{"unknown_field_rejected", map[string]any{"unknownCredential": "injected", "password": "replacement-password"}, "Unsupported credential field"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newSessionValidateFixture(t)
			if err := pluginsystem.ApplyAuthPatch(f.instance, f.manifest, map[string]any{"password": "valid-stored-password"}); err != nil {
				t.Fatal(err)
			}
			if err := f.app.Save(f.instance); err != nil {
				t.Fatal(err)
			}
			before := f.instance.GetString("auth")
			request := f.body()
			request["auth"] = tc.patch
			status, body := f.request(t, request)
			if status != 400 || !strings.Contains(body, tc.message) || f.runtime.calls.Load() != 0 || f.providerRequests.Load() != 0 {
				t.Fatalf("invalid candidate reached validation: %d %s", status, body)
			}
			fresh, err := f.app.FindRecordById("plugin_instances", f.instance.Id)
			if err != nil || fresh.GetString("auth") != before {
				t.Fatal("validation persisted its candidate")
			}
		})
	}
}

func TestValidateRejectsConnectionChangeAfterProviderResult(t *testing.T) {
	f := newSessionValidateFixture(t)
	arrived, resume := make(chan struct{}), make(chan struct{})
	cleanupConnectionBarrier(t, resume)
	f.runtime.beforeReturn = func() { close(arrived); <-resume }
	type response struct {
		status int
		body   string
	}
	result := make(chan response, 1)
	go func() { status, body := f.request(t, f.body()); result <- response{status, body} }()
	waitConnectionBarrier(t, arrived)
	_, err := pluginsystem.PatchConnection(f.app, pluginsystem.CaptureConnection(f.instance), false, func(fresh *core.Record) error { pluginsystem.ResetConnection(fresh); return nil })
	if err != nil {
		t.Fatal(err)
	}
	close(resume)
	select {
	case response := <-result:
		if response.status != 409 || !strings.Contains(response.body, `"code":"plugin_connection_conflict"`) {
			t.Fatalf("stale Validate result: %#v", response)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Validate stuck")
	}
	if f.providerRequests.Load() != 1 {
		t.Fatal("Validate replayed provider call")
	}
}

func TestSendBusyUsesSharedInstanceLockWithoutRuntimeWork(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	router, err := newConnectionSendRouter(f)
	if err != nil {
		t.Fatal(err)
	}
	release, acquired := pluginsystem.TryLockInstance(f.instanceID)
	if !acquired {
		t.Fatal("test lock unavailable")
	}
	defer release()
	data, _ := json.Marshal(map[string]any{"pluginId": "strava", "trailId": "not-read-while-busy"})
	request := httptest.NewRequest(http.MethodPost, "/send", bytes.NewReader(data))
	request.Header.Set("Authorization", f.authToken)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 409 || !strings.Contains(response.Body.String(), `"code":"plugin_instance_busy"`) || f.tokenRequests.Load() != 0 {
		t.Fatalf("busy send: %d %s", response.Code, response.Body.String())
	}
}

func newConnectionSendRouter(f *oauthMetadataFixture) (http.Handler, error) {
	// Runtime must never be resolved while another action holds the lock.
	return buildConnectionRoute(f.app, "/send", func(e *core.RequestEvent) error {
		return pluginSystemTrailSend(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) { panic("busy Send reached runtime") })
	})
}

func buildConnectionRoute(app core.App, path string, handler func(*core.RequestEvent) error) (http.Handler, error) {
	router, err := apis.NewRouter(app)
	if err != nil {
		return nil, err
	}
	router.POST(path, handler)
	return router.BuildMux()
}

func TestSessionRefreshDiscardsObsoleteToken(t *testing.T) {
	f := newSessionValidateFixture(t)
	// Session refresh is an existing connection action, so use explicit enabled
	// state in its schema and record rather than bypassing the C+ guard.
	collection := f.instance.Collection()
	collection.Fields.Add(&core.BoolField{Name: "enabled"})
	if err := f.app.Save(collection); err != nil {
		t.Fatal(err)
	}
	f.instance.Set("enabled", true)
	if err := f.app.Save(f.instance); err != nil {
		t.Fatal(err)
	}
	arrived, resume := make(chan struct{}), make(chan struct{})
	cleanupConnectionBarrier(t, resume)
	f.runtime.beforeReturn = func() { close(arrived); <-resume }
	auth := map[string]any{"email": "stored@example.test", "password": "replacement-password"}
	plugin := pluginsystem.LocalPlugin{Manifest: f.manifest}
	config := effectivePluginConfig(f.app, f.manifest.ID, f.instance)
	pluginsystem.MergePluginConfig(config, map[string]any{"host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": f.baseURL, "basePath": "/draft"}}}})
	policy, err := pluginInstancePolicy(plugin, config)
	if err != nil {
		t.Fatal(err)
	}
	spec := &pluginsystem.HostRequestSpec{Auth: "dashboard"}
	result := make(chan error, 1)
	go func() {
		result <- pluginsystem.InjectHostRequestAuth(context.Background(), pluginsystem.AuthInjectionInput{App: f.app, Plugin: plugin, Instance: f.instance, Runtime: f.runtime, Auth: auth, Config: pluginRuntimeConfig(config), Policy: policy.WithHostAuth(auth), Spec: spec})
	}()
	waitConnectionBarrier(t, arrived)
	_, err = pluginsystem.PatchConnection(f.app, pluginsystem.CaptureConnection(f.instance), false, func(fresh *core.Record) error { fresh.Set("enabled", false); return nil })
	if err != nil {
		t.Fatal(err)
	}
	close(resume)
	select {
	case err := <-result:
		if err != pluginsystem.ErrInstanceDisabled {
			t.Fatalf("obsolete session: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("session stuck")
	}
	if len(spec.Headers) != 0 {
		t.Fatal("obsolete session token was injected")
	}
}

func TestInstanceAuthInjectionRejectsEveryInactiveContextBeforeProviderIO(t *testing.T) {
	for _, authType := range []string{pluginsystem.AuthTypeAPIKey, pluginsystem.AuthTypeBearer, pluginsystem.AuthTypeSession, pluginsystem.AuthTypeOAuth2} {
		t.Run(authType, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			instance, _ := f.storedAuth(t)
			manifest := f.plugin.Manifest
			manifest.Auth.Contexts["inactive"] = pluginsystem.AuthContext{Type: authType, SecretField: "inactiveToken", SecretFields: []string{"inactiveToken"}, Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModePlugin, Function: "must_not_call"}}
			manifest.Permissions.Auth = append(manifest.Permissions.Auth, "inactive")
			spec := &pluginsystem.HostRequestSpec{Auth: "inactive"}
			err := pluginsystem.InjectHostRequestAuth(context.Background(), pluginsystem.AuthInjectionInput{App: f.app, Plugin: pluginsystem.LocalPlugin{Manifest: manifest}, Instance: instance, Auth: map[string]any{"inactiveToken": "inactive-opaque-credential"}, Spec: spec})
			if err == nil || len(spec.Headers) != 0 || len(spec.Target.Query) != 0 || f.tokenRequests.Load() != 0 {
				t.Fatal("inactive context injected credentials or reached provider")
			}
		})
	}
}
