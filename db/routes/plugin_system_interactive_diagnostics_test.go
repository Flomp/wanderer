package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"pocketbase/pluginsystem"
)

// Capture the same App logger the production routes use, while leaving real
// PocketBase auth, owned records, hooks and provider HTTP boundaries in place.
type interactiveDiagnosticApp struct {
	core.App
	logger *slog.Logger
}

func (a interactiveDiagnosticApp) Logger() *slog.Logger { return a.logger }

func interactiveDiagnosticHandler(t *testing.T, app core.App, output *bytes.Buffer, paths map[string]func(*core.RequestEvent) error) http.Handler {
	t.Helper()
	// pbtests keeps MaxDays=0 to disable asynchronous request logs. This helper
	// captures synchronous route diagnostics; persisted _logs has separate tests.
	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(output, nil))
	for path, handler := range paths {
		router.POST(path, func(e *core.RequestEvent) error {
			e.App = interactiveDiagnosticApp{App: e.App, logger: logger}
			return handler(e)
		})
	}
	handler, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

// SetDefault also redirects the standard logger and resets its flags.
// Restore all global state, including caller-defined prefixes, after each test.
func setInteractiveDefaultLogger(t *testing.T, output *bytes.Buffer) {
	t.Helper()
	previousLogger := slog.Default()
	previousWriter, previousFlags, previousPrefix := log.Writer(), log.Flags(), log.Prefix()
	slog.SetDefault(slog.New(slog.NewJSONHandler(output, nil)))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
		log.SetPrefix(previousPrefix)
	})
}

func interactiveDiagnosticRequest(t *testing.T, handler http.Handler, token, path string, body map[string]any) (int, map[string]any) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid API response: %s", response.Body.String())
	}
	return response.Code, result
}

func assertInteractiveDiagnostic(t *testing.T, output *bytes.Buffer, want map[string]any, secrets ...string) {
	t.Helper()
	// A single route boundary logs once, including configuration failures.
	lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
	if len(lines) != 1 {
		t.Fatalf("expected one App diagnostic, got %s", output.String())
	}
	var record map[string]any
	if err := json.Unmarshal(lines[0], &record); err != nil {
		t.Fatal(err)
	}
	for key, value := range want {
		if record[key] != value {
			t.Errorf("diagnostic %s=%#v, want %#v: %s", key, record[key], value, output.String())
		}
	}
	for _, secret := range secrets {
		if secret != "" && strings.Contains(output.String(), secret) {
			t.Errorf("diagnostic leaked %q: %s", secret, output.String())
		}
	}
}

func TestOAuthCallbackLogsProviderClassificationWithoutSecrets(t *testing.T) {
	for _, tc := range []struct {
		name             string
		status           int
		rfc, code, class string
	}{
		{"invalid grant", 400, "invalid_grant", "invalid_grant", "authentication"},
		{"invalid client", 401, "invalid_client", "invalid_client", "authentication"},
		{"rate limit", 429, "", "rate_limited", "rate_limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				f.tokenRequests.Add(1)
				w.Header().Set("Content-Type", "application/json")
				if tc.status == 429 {
					w.Header().Set("Retry-After", "0")
				}
				w.WriteHeader(tc.status)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": tc.rfc, "error_description": "provider-body-secret"})
			}))
			t.Cleanup(provider.Close)
			authContext := f.plugin.Manifest.Auth.Contexts["oauth_access_token"]
			authContext.TokenURL = provider.URL + "/oauth/token?token=query-secret"
			f.plugin.Manifest.Auth.Contexts["oauth_access_token"] = authContext
			for i := range f.plugin.Manifest.Permissions.Network.Connectors {
				if f.plugin.Manifest.Permissions.Network.Connectors[i].Name == "oauth" {
					f.plugin.Manifest.Permissions.Network.Connectors[i].FixedBaseURL = provider.URL + "/oauth"
				}
			}
			installed, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", "strava")
			if err != nil {
				t.Fatal(err)
			}
			installed.Set("manifest", f.plugin.Manifest)
			if err := f.app.Save(installed); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			f.handler = interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{
				"/api/plugin-system/oauth/start":    PluginSystemOAuthStart,
				"/api/plugin-system/oauth/callback": PluginSystemOAuthCallback,
			})
			started := f.request(t, http.MethodPost, "/api/plugin-system/oauth/start", map[string]any{"pluginId": "strava", "instanceId": f.instanceID, "expected_connection_generation": oauthFixtureGeneration(t, f), "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback"})
			state := started["state"].(string)
			status, body := interactiveDiagnosticRequest(t, f.handler, f.authToken, "/api/plugin-system/oauth/callback", map[string]any{"instanceId": f.instanceID, "state": state, "code": "test-authorization-code"})
			if status != http.StatusBadRequest {
				t.Fatalf("callback status %d: %#v", status, body)
			}
			operation := body["data"].(map[string]any)["operation"].(map[string]any)
			if operation["code"] != tc.code {
				t.Fatalf("API code changed: %#v", body)
			}
			instance, err := f.app.FindRecordById("plugin_instances", f.instanceID)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]any{"action": "oauth_callback", "owner": instance.GetString("user"), "plugin": "strava", "instance": f.instanceID, "code": tc.code, "class": tc.class, "operation": pluginsystem.OperationOAuthExchange, "http_status": float64(tc.status)}
			if tc.status == 429 {
				want["retry_after_seconds"] = float64(0)
			}
			assertInteractiveDiagnostic(t, &output, want, provider.URL, "query-secret", "provider-body-secret", f.expectedClientSecret, f.expectedRefreshToken, "strava-test-access", state, "test-authorization-code")
			if f.tokenRequests.Load() != 1 {
				t.Fatal("callback did not exchange the code exactly once")
			}
		})
	}
}

func TestOAuthStartAndRevokeLogSafeLocalFailures(t *testing.T) {
	t.Run("start configuration", func(t *testing.T) {
		f := newOAuthMetadataFixture(t)
		var output bytes.Buffer
		handler := interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{"/api/test/start": PluginSystemOAuthStart})
		status, _ := interactiveDiagnosticRequest(t, handler, f.authToken, "/api/test/start", map[string]any{"pluginId": "strava", "instanceId": f.instanceID, "expected_connection_generation": oauthFixtureGeneration(t, f), "authContext": "context-secret", "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback"})
		if status != 400 {
			t.Fatalf("status %d", status)
		}
		assertInteractiveDiagnostic(t, &output, map[string]any{"action": "oauth_start", "plugin": "strava", "code": "invalid_configuration", "class": "configuration", "operation": pluginsystem.OperationOAuthExchange}, "request-instance-secret", "context-secret")
	})
	t.Run("revoke storage failure", func(t *testing.T) {
		f := newOAuthMetadataFixture(t)
		f.app.OnRecordUpdate("plugin_instances").BindFunc(func(e *core.RecordEvent) error { return errors.New("database-write-secret") })
		var output bytes.Buffer
		handler := interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{"/api/test/revoke": PluginSystemOAuthRevoke})
		status, _ := interactiveDiagnosticRequest(t, handler, f.authToken, "/api/test/revoke", map[string]any{"instanceId": f.instanceID, "expected_connection_generation": oauthFixtureGeneration(t, f)})
		if status == 200 {
			t.Fatal("failed save unexpectedly succeeded")
		}
		assertInteractiveDiagnostic(t, &output, map[string]any{"action": "oauth_revoke", "instance": f.instanceID, "code": "internal_error", "class": "internal", "cause": "database_save", "operation": "oauth_revoke"}, "database-write-secret", f.expectedClientSecret, f.expectedRefreshToken)
	})
}

// This fixture uses only valid persisted settings and shared route schemas;
// the diagnostic tests do not depend on B0's unsaved-config or active-auth UX.
type interactivePluginFixture struct {
	app             *pbtests.TestApp
	owner, instance *core.Record
	manifest        pluginsystem.Manifest
	baseURL         string
	providerClient  *http.Client
}

func newInteractivePluginFixture(t *testing.T, providerHandler http.HandlerFunc) *interactivePluginFixture {
	t.Helper()
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", oauthMetadataTestKey)
	app, err := pbtests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	// Most route-boundary tests never contact a provider. A valid configured
	// origin is sufficient; do not bind a listener merely to create records.
	f := &interactivePluginFixture{app: app, baseURL: "https://provider.example.test"}
	if providerHandler != nil {
		provider := httptest.NewServer(providerHandler)
		t.Cleanup(provider.Close)
		f.baseURL = provider.URL
		f.providerClient = provider.Client()
		f.providerClient.Timeout = 5 * time.Second
	}
	users := core.NewAuthCollection("interactive_diagnostic_users")
	if err := app.Save(users); err != nil {
		t.Fatal(err)
	}
	f.owner = core.NewRecord(users)
	f.owner.SetEmail("interactive@example.test")
	f.owner.SetPassword("test-owner-password")
	if err := app.Save(f.owner); err != nil {
		t.Fatal(err)
	}
	f.manifest = pluginsystem.Manifest{ManifestVersion: "1.0", ID: "interactive-test", Type: pluginsystem.PluginTypeTrails, Name: "Interactive fixture", Version: "0.1.0", Runtime: pluginsystem.RuntimeManifest{Type: pluginsystem.RuntimeWASM, Entrypoint: "plugin.wasm"}, Capabilities: []pluginsystem.CapabilityManifest{{Name: "list_routes", Version: "v1", Export: "list_routes_v1"}, {Name: "prepare_trail_send", Version: "v1", Export: "prepare_trail_send_v1"}}, Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{"dashboard": {Type: pluginsystem.AuthTypeSession, Fields: []string{"email", "password"}, SecretFields: []string{"password"}, Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModePlugin, Function: "refresh_auth_v1"}}}}, Permissions: pluginsystem.PermissionManifest{Auth: []string{"dashboard"}, Network: pluginsystem.NetworkPermissions{Connectors: []pluginsystem.ConnectorTargetPermission{{Name: "api", Type: pluginsystem.ConnectorTypeConfigured, ConfigKey: "server", AllowedPathPrefixes: []string{"/"}, Auth: []string{"dashboard"}}}}, Downloads: pluginsystem.DownloadPermissions{MaxBytes: 1024, ContentTypes: []string{"application/json"}}}}
	installed := core.NewBaseCollection("installed_plugins")
	installed.Fields.Add(&core.TextField{Name: "plugin_id"}, &core.TextField{Name: "path"}, &core.JSONField{Name: "manifest"}, &core.JSONField{Name: "config"})
	if err := app.Save(installed); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "plugin.wasm"), []byte("\x00asm\x01\x00\x00\x00"), 0600); err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(installed)
	record.Set("plugin_id", f.manifest.ID)
	record.Set("path", directory)
	record.Set("manifest", f.manifest)
	record.Set("config", map[string]any{"host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": f.baseURL, "allowPrivate": true}}}})
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	instances := core.NewBaseCollection("plugin_instances")
	instances.Fields.Add(&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1}, &core.TextField{Name: "plugin_id"}, &core.BoolField{Name: "enabled"}, &core.JSONField{Name: "auth"}, &core.JSONField{Name: "config"}, &core.TextField{Name: "auth_context"}, &core.TextField{Name: "connection_generation"}, &core.TextField{Name: "oauth_token_revision"})
	if err := app.Save(instances); err != nil {
		t.Fatal(err)
	}
	f.instance = core.NewRecord(instances)
	f.instance.Set("user", f.owner.Id)
	f.instance.Set("plugin_id", f.manifest.ID)
	f.instance.Set("enabled", true)
	f.instance.Set("auth_context", "dashboard")
	f.instance.Set("connection_generation", pluginsystem.NewConnectionRevision())
	f.instance.Set("oauth_token_revision", pluginsystem.NewConnectionRevision())
	f.instance.Set("auth", map[string]any{"email": "owner@example.test", "password": "credential-secret", "oauthContext": "dashboard"})
	if err := app.Save(f.instance); err != nil {
		t.Fatal(err)
	}
	return f
}

type diagnosticValidateRuntime struct{ guestError error }

func (r diagnosticValidateRuntime) Call(context.Context, pluginsystem.LocalPlugin, string, []byte, pluginsystem.RequestPolicyContext) ([]byte, error) {
	// Test the actual route boundary for a typed guest failure. Provider HTTP
	// and candidate-config policy resolution have independent regressions.
	return nil, r.guestError
}
func (diagnosticValidateRuntime) OpenSession(context.Context, pluginsystem.LocalPlugin, pluginsystem.RequestPolicyContext) (pluginsystem.RuntimeSession, error) {
	return nil, errors.New("unexpected session open")
}

func TestSessionValidateLogsGuestMetadataAtAppBoundary(t *testing.T) {
	f := newInteractivePluginFixture(t, nil)
	zero := 0
	runtime := diagnosticValidateRuntime{guestError: fmt.Errorf("provider-body-secret: %w", pluginsystem.PluginFailure{PluginError: pluginsystem.PluginError{Code: "rate_limited", Message: "provider-token-secret", Operation: pluginsystem.OperationLogin, HTTPStatus: 429, RetryAfterSeconds: &zero}})}
	var output bytes.Buffer
	handler := interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{"/api/test/validate": func(e *core.RequestEvent) error {
		return pluginSystemSessionAuthValidate(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) { return runtime, nil })
	}})
	token, err := f.owner.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	status, body := interactiveDiagnosticRequest(t, handler, token, "/api/test/validate", map[string]any{"pluginId": f.manifest.ID, "instanceId": f.instance.Id, "expected_connection_generation": f.instance.GetString("connection_generation")})
	if status != 400 {
		t.Fatalf("validate status %d: %#v", status, body)
	}
	assertInteractiveDiagnostic(t, &output, map[string]any{"action": "session_validate", "owner": f.owner.Id, "plugin": f.manifest.ID, "instance": f.instance.Id, "code": "rate_limited", "class": "rate_limit", "operation": pluginsystem.OperationLogin, "http_status": float64(429), "retry_after_seconds": float64(0)}, "provider-body-secret", "provider-token-secret", f.baseURL, "credential-secret")
}

// The guest supplies a send plan. Tests below cover the actual Send route and
// classifier, with either a local test transport or the productive executor.
type diagnosticSendRuntime struct{ request pluginsystem.HostRequestSpec }

func (*diagnosticSendRuntime) Call(context.Context, pluginsystem.LocalPlugin, string, []byte, pluginsystem.RequestPolicyContext) ([]byte, error) {
	return nil, errors.New("unexpected runtime call")
}
func (r *diagnosticSendRuntime) OpenSession(context.Context, pluginsystem.LocalPlugin, pluginsystem.RequestPolicyContext) (pluginsystem.RuntimeSession, error) {
	return diagnosticSendSession{request: r.request}, nil
}

type diagnosticSendSession struct{ request pluginsystem.HostRequestSpec }

func (s diagnosticSendSession) Call(context.Context, string, []byte) ([]byte, error) {
	return json.Marshal(pluginsystem.TrailSendPlan{Request: s.request})
}
func (diagnosticSendSession) Close(context.Context) error { return nil }

func TestSendLogsRealProviderRejectionAtAppBoundary(t *testing.T) {
	var providerRequests atomic.Int32
	f := newInteractivePluginFixture(t, func(w http.ResponseWriter, r *http.Request) {
		providerRequests.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/reject" || r.URL.Query().Get("token") != "query-secret" || r.Header.Get("Authorization") != "Bearer provider-token-secret" {
			t.Errorf("unexpected provider request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"error":"provider-body-secret"}`))
	})
	trail := newInteractiveDiagnosticTrail(t, f)
	runtime := &diagnosticSendRuntime{request: pluginsystem.HostRequestSpec{Method: http.MethodGet, Target: pluginsystem.RequestTarget{Type: "connector", Connector: "api", Path: "/reject", Query: []pluginsystem.QueryParam{{Name: "token", Value: "query-secret"}}}, Headers: map[string]string{"Authorization": "Bearer provider-token-secret"}}}
	var output bytes.Buffer
	// This seam test checks real provider HTTP, production classification and
	// the route log/API. Productive executor duplicate logging is tested below.
	setInteractiveDefaultLogger(t, &output)
	handler := interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{"/api/test/send": func(e *core.RequestEvent) error {
		return pluginSystemTrailSendWithHostRequest(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) { return runtime, nil }, func(ctx context.Context, manifest pluginsystem.Manifest, policy pluginsystem.RequestPolicyContext, spec pluginsystem.HostRequestSpec, _ pluginsystem.HostRequestOptions) (pluginsystem.HostResponse, error) {
			// Exercise the real connector target resolution, HTTP response and
			// production Send classification/API boundary. Only the transport is
			// local to this test; production's deliberate loopback block remains.
			resolved, err := pluginsystem.ValidateAndResolveHostRequestSpec(manifest, spec, policy)
			if err != nil {
				return pluginsystem.HostResponse{}, err
			}
			if resolved.Connector.BaseURL != f.baseURL {
				return pluginsystem.HostResponse{}, errors.New("unexpected test provider origin")
			}
			request, err := http.NewRequestWithContext(ctx, spec.Method, resolved.URL.String(), nil)
			if err != nil {
				return pluginsystem.HostResponse{}, err
			}
			for key, value := range spec.Headers {
				request.Header.Set(key, value)
			}
			response, err := f.providerClient.Do(request)
			if err != nil {
				return pluginsystem.HostResponse{}, err
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			return pluginsystem.HostResponse{Status: response.StatusCode, HeaderValues: response.Header, Body: body}, err
		})
	}})
	token, err := f.owner.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	status, body := interactiveDiagnosticRequest(t, handler, token, "/api/test/send", map[string]any{"pluginId": f.manifest.ID, "trailId": trail.Id})
	if status != 400 {
		t.Fatalf("send status %d: %#v", status, body)
	}
	operation := body["data"].(map[string]any)["operation"].(map[string]any)
	params := operation["params"].(map[string]any)
	if operation["code"] != "rate_limited" || params["operation"] != pluginsystem.OperationSend || params["httpStatus"] != float64(429) || params["retryAfterSeconds"] != float64(0) {
		t.Fatalf("send API lost provider classification/zero delay: %#v", body)
	}
	if providerRequests.Load() != 1 {
		t.Fatalf("provider requests %d, want exactly one", providerRequests.Load())
	}
	assertInteractiveDiagnostic(t, &output, map[string]any{"action": "send", "owner": f.owner.Id, "plugin": f.manifest.ID, "instance": f.instance.Id, "code": "rate_limited", "class": "rate_limit", "operation": pluginsystem.OperationSend, "http_status": float64(429), "retry_after_seconds": float64(0)}, f.baseURL, "query-secret", "provider-token-secret", "provider-body-secret", "credential-secret", "guest-trail-name-secret", "source-file-secret.gpx")
}

func newInteractiveDiagnosticTrail(t *testing.T, f *interactivePluginFixture) *core.Record {
	t.Helper()
	trails := core.NewBaseCollection("trails")
	trails.Fields.Add(&core.BoolField{Name: "public"}, &core.TextField{Name: "name"}, &core.FileField{Name: "gpx", MaxSelect: 1})
	if err := f.app.Save(trails); err != nil {
		t.Fatal(err)
	}
	trail := core.NewRecord(trails)
	trail.Set("public", true)
	trail.Set("name", "guest-trail-name-secret")
	file, err := filesystem.NewFileFromBytes([]byte(`<gpx version="1.1"></gpx>`), "source-file-secret.gpx")
	if err != nil {
		t.Fatal(err)
	}
	trail.Set("gpx", file)
	if err := f.app.Save(trail); err != nil {
		t.Fatal(err)
	}
	return trail
}

func TestSendProductiveExecutorLogsOnceAtRouteBoundary(t *testing.T) {
	f := newInteractivePluginFixture(t, nil)
	// The configured origin is syntactically valid. Production's connector
	// client rejects literal loopback before dial, including allowPrivate=true.
	// No listener, private interface, external DNS or injected executor is used.
	f.baseURL = "http://127.0.0.1:1"
	installed, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", f.manifest.ID)
	if err != nil {
		t.Fatal(err)
	}
	installed.Set("config", map[string]any{"host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": f.baseURL, "allowPrivate": true}}}})
	if err := f.app.Save(installed); err != nil {
		t.Fatal(err)
	}
	trail := newInteractiveDiagnosticTrail(t, f)
	runtime := &diagnosticSendRuntime{request: pluginsystem.HostRequestSpec{Method: http.MethodGet, Target: pluginsystem.RequestTarget{Type: "connector", Connector: "api", Path: "/reject", Query: []pluginsystem.QueryParam{{Name: "token", Value: "query-secret"}}}, Headers: map[string]string{"Authorization": "Bearer provider-token-secret"}}}
	var output bytes.Buffer
	// Capture both App and global diagnostics. The actual executor used to log
	// once itself and once at this route, so the one-line assertion is decisive.
	setInteractiveDefaultLogger(t, &output)
	handler := interactiveDiagnosticHandler(t, f.app, &output, map[string]func(*core.RequestEvent) error{"/api/test/send": func(e *core.RequestEvent) error {
		return pluginSystemTrailSend(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) { return runtime, nil })
	}})
	token, err := f.owner.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	status, body := interactiveDiagnosticRequest(t, handler, token, "/api/test/send", map[string]any{"pluginId": f.manifest.ID, "trailId": trail.Id})
	if status != http.StatusBadRequest {
		t.Fatalf("send status %d: %#v", status, body)
	}
	operation := body["data"].(map[string]any)["operation"].(map[string]any)
	if operation["code"] != "provider_unavailable" || operation["params"].(map[string]any)["operation"] != pluginsystem.OperationHostHTTP {
		t.Fatalf("send did not reach productive connector transport: %#v", body)
	}
	assertInteractiveDiagnostic(t, &output, map[string]any{"action": "send", "owner": f.owner.Id, "plugin": f.manifest.ID, "instance": f.instance.Id, "code": "provider_unavailable", "operation": pluginsystem.OperationHostHTTP, "cause": "network_unavailable"}, f.baseURL, "query-secret", "provider-token-secret", "credential-secret", "guest-trail-name-secret", "source-file-secret.gpx")
}

func oauthFixtureGeneration(t *testing.T, fixture *oauthMetadataFixture) string {
	t.Helper()
	instance, err := fixture.app.FindRecordById("plugin_instances", fixture.instanceID)
	if err != nil {
		t.Fatal(err)
	}
	return instance.GetString("connection_generation")
}
