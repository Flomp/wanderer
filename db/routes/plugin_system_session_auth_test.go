package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	pbtests "github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/security"
	"pocketbase/pluginsystem"
)

// The guest seam captures the actual route input and delegates its connector
// request to the production Host HTTP implementation, including IP policy.
type sessionValidateRuntime struct {
	t          *testing.T
	calls      atomic.Int32
	input      pluginSessionAuthRefreshInput
	policy     pluginsystem.RequestPolicyContext
	guestError error
}

func (r *sessionValidateRuntime) Call(ctx context.Context, plugin pluginsystem.LocalPlugin, export string, input []byte, policy pluginsystem.RequestPolicyContext) ([]byte, error) {
	r.calls.Add(1)
	if export != "refresh_auth_v1" {
		r.t.Errorf("wrong refresh export %q", export)
	}
	if err := json.Unmarshal(input, &r.input); err != nil {
		return nil, err
	}
	r.policy = policy
	response, err := pluginsystem.ExecuteHostRequest(ctx, plugin.Manifest, policy, pluginsystem.HostRequestSpec{Method: http.MethodGet, Target: pluginsystem.RequestTarget{Type: "connector", Connector: "api", Path: "/login"}}, pluginsystem.HostRequestOptions{})
	if err != nil {
		return nil, err
	}
	if r.guestError != nil {
		return nil, r.guestError
	}
	return response.Body, nil
}
func (r *sessionValidateRuntime) OpenSession(context.Context, pluginsystem.LocalPlugin, pluginsystem.RequestPolicyContext) (pluginsystem.RuntimeSession, error) {
	return nil, fmt.Errorf("unexpected session open")
}

type sessionValidateFixture struct {
	app              *pbtests.TestApp
	owner            *core.Record
	instance         *core.Record
	manifest         pluginsystem.Manifest
	handler          http.Handler
	runtime          *sessionValidateRuntime
	providerRequests atomic.Int32
	baseURL          string
}

func newSessionValidateFixture(t *testing.T) *sessionValidateFixture {
	t.Helper()
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", oauthMetadataTestKey)
	app, err := pbtests.NewTestApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	f := &sessionValidateFixture{app: app, runtime: &sessionValidateRuntime{t: t}}
	// Loopback is forbidden even with allowPrivate. Bind an actual RFC1918
	// address to exercise the same policy used by a local Immich-style connector.
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatal(err)
	}
	var listener net.Listener
	for _, address := range addresses {
		network, ok := address.(*net.IPNet)
		if !ok || network.IP.To4() == nil || !network.IP.IsPrivate() || network.IP.IsLoopback() {
			continue
		}
		listener, err = net.Listen("tcp", net.JoinHostPort(network.IP.String(), "0"))
		if err == nil {
			break
		}
	}
	if listener == nil {
		t.Fatal("a private network interface is required for the configured connector regression")
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		f.providerRequests.Add(1)
		if request.URL.Path != "/draft/login" {
			t.Errorf("wrong configured base path %q", request.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"provider-session"}`))
	}))
	server.Listener.Close()
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)
	f.baseURL = server.URL
	users := core.NewAuthCollection("session_validation_users")
	if err := app.Save(users); err != nil {
		t.Fatal(err)
	}
	f.owner = core.NewRecord(users)
	f.owner.SetEmail("session-validation@example.test")
	f.owner.SetPassword("session-test-password")
	if err := app.Save(f.owner); err != nil {
		t.Fatal(err)
	}
	installed := core.NewBaseCollection("installed_plugins")
	installed.Fields.Add(&core.TextField{Name: "plugin_id"}, &core.TextField{Name: "path"}, &core.JSONField{Name: "manifest"}, &core.JSONField{Name: "config"})
	if err := app.Save(installed); err != nil {
		t.Fatal(err)
	}
	f.manifest = pluginsystem.Manifest{ManifestVersion: "1.0", ID: "session-validation", Type: pluginsystem.PluginTypeTrails, Name: "Session validation fixture", Version: "0.1.0", Runtime: pluginsystem.RuntimeManifest{Type: pluginsystem.RuntimeWASM, Entrypoint: "plugin.wasm"}, Capabilities: []pluginsystem.CapabilityManifest{{Name: "list_routes", Version: "v1", Export: "list_routes_v1"}}, Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{"dashboard": {Type: pluginsystem.AuthTypeSession, Fields: []string{"email", "password"}, SecretFields: []string{"password"}, Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModePlugin, Function: "refresh_auth_v1"}}, "inactive": {Type: pluginsystem.AuthTypeAPIKey, SecretField: "apiKey"}}}, Permissions: pluginsystem.PermissionManifest{Auth: []string{"dashboard", "inactive"}, Network: pluginsystem.NetworkPermissions{Connectors: []pluginsystem.ConnectorTargetPermission{{Name: "api", Type: pluginsystem.ConnectorTypeConfigured, ConfigKey: "server", SupportsStorageRedirects: true, AllowedPathPrefixes: []string{"/"}, Auth: []string{"dashboard"}}}}, Downloads: pluginsystem.DownloadPermissions{MaxBytes: 1024, ContentTypes: []string{"application/json"}}}, ConfigSchema: []pluginsystem.ConfigField{{Key: "limit", Type: "number"}, {Key: "after", Type: "date"}}}
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "plugin.wasm"), []byte("\x00asm\x01\x00\x00\x00"), 0600); err != nil {
		t.Fatal(err)
	}
	installedRecord := core.NewRecord(installed)
	installedRecord.Set("plugin_id", f.manifest.ID)
	installedRecord.Set("path", directory)
	installedRecord.Set("manifest", f.manifest)
	installedRecord.Set("config", map[string]any{"plugin": map[string]any{"limit": 1, "after": "2026-01-01"}, "host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": "https://installed.invalid", "allowPrivate": true}}}})
	if err := app.Save(installedRecord); err != nil {
		t.Fatal(err)
	}
	instances := core.NewBaseCollection("plugin_instances")
	instances.Fields.Add(&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1}, &core.TextField{Name: "plugin_id"}, &core.JSONField{Name: "auth"}, &core.JSONField{Name: "config"})
	if err := app.Save(instances); err != nil {
		t.Fatal(err)
	}
	broken, err := security.Encrypt([]byte("old-password"), "ffffffffffffffffffffffffffffffff")
	if err != nil {
		t.Fatal(err)
	}
	f.instance = core.NewRecord(instances)
	f.instance.Set("user", f.owner.Id)
	f.instance.Set("plugin_id", f.manifest.ID)
	f.instance.Set("auth", map[string]any{"email": "stored@example.test", "password": broken, "oauthContext": "dashboard", "apiKey": broken, "oauthState": broken})
	f.instance.Set("config", map[string]any{"plugin": map[string]any{"limit": 2}, "host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": "https://stored.invalid"}}}})
	if err := app.Save(f.instance); err != nil {
		t.Fatal(err)
	}
	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	router.POST("/api/plugins/auth/validate", func(e *core.RequestEvent) error {
		return pluginSystemSessionAuthValidate(e, func(pluginsystem.LocalPlugin) (pluginsystem.Runtime, error) { return f.runtime, nil })
	})
	f.handler, err = router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *sessionValidateFixture) request(t *testing.T, body map[string]any) (int, string) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	token, err := f.owner.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/plugins/auth/validate", bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", token)
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, request)
	return response.Code, response.Body.String()
}
func (f *sessionValidateFixture) body() map[string]any {
	return map[string]any{"pluginId": f.manifest.ID, "instanceId": f.instance.Id, "auth": map[string]any{"password": "replacement-password"}, "config": map[string]any{"plugin": map[string]any{"limit": 3}, "host": map[string]any{"connectors": map[string]any{"server": map[string]any{"baseURL": f.baseURL, "basePath": "/draft"}}}}}
}

func TestSessionValidateUsesDraftConfigAndReplacementCredentialsWithoutSaving(t *testing.T) {
	f := newSessionValidateFixture(t)
	beforeAuth, beforeConfig := f.instance.GetString("auth"), f.instance.GetString("config")
	status, body := f.request(t, f.body())
	if status != http.StatusOK {
		t.Fatalf("Validate status %d: %s", status, body)
	}
	if f.runtime.calls.Load() != 1 || f.providerRequests.Load() != 1 {
		t.Fatal("validation did not reach the intended provider exactly once")
	}
	if f.runtime.input.Auth["email"] != "stored@example.test" || f.runtime.input.Auth["password"] != "replacement-password" || len(f.runtime.input.Auth) != 2 {
		t.Fatalf("incorrect active auth projection: %#v", f.runtime.input.Auth)
	}
	if f.runtime.input.Config["limit"] != float64(3) || f.runtime.input.Config["after"] != "2026-01-01" || len(f.runtime.input.Config) != 2 {
		t.Fatalf("incorrect effective plugin config: %#v", f.runtime.input.Config)
	}
	if !f.runtime.policy.Connectors["api"].AllowPrivate {
		t.Fatal("allowPrivate was silently restricted")
	}
	fresh, err := f.app.FindRecordById("plugin_instances", f.instance.Id)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.GetString("auth") != beforeAuth || fresh.GetString("config") != beforeConfig {
		t.Fatal("validation persisted draft config or credentials")
	}
}

func TestSessionValidateRejectsInvalidDraftBeforeProvider(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(map[string]any, map[string]any)
	}{
		{"null root config", func(body map[string]any, _ map[string]any) { body["config"] = nil }},
		{"invalid plugin section", func(body map[string]any, _ map[string]any) { body["config"].(map[string]any)["plugin"] = []any{} }},
		{"invalid host section", func(body map[string]any, _ map[string]any) { body["config"].(map[string]any)["host"] = nil }},
		{"missing connector base", func(_ map[string]any, c map[string]any) { c["baseURL"] = "" }},
		{"custom tls forbidden", func(_ map[string]any, c map[string]any) {
			c["tls"] = map[string]any{"mode": "customCA", "caBundle": "secret-invalid-ca"}
		}},
		{"unknown tls mode", func(_ map[string]any, c map[string]any) { c["tls"] = map[string]any{"mode": "disabled"} }},
		{"wrong private flag type", func(_ map[string]any, c map[string]any) { c["allowPrivate"] = "true" }},
		{"invalid storage origin", func(_ map[string]any, c map[string]any) {
			c["storageOrigins"] = map[string]any{"archive": map[string]any{"baseURL": "file:///private-secret"}}
		}},
		{"unreadable active credential", func(body map[string]any, _ map[string]any) { body["auth"] = map[string]any{"password": ""} }},
		{"unknown explicit context", func(body map[string]any, _ map[string]any) { body["authContext"] = "missing" }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f := newSessionValidateFixture(t)
			body := f.body()
			connector := body["config"].(map[string]any)["host"].(map[string]any)["connectors"].(map[string]any)["server"].(map[string]any)
			test.mutate(body, connector)
			status, response := f.request(t, body)
			if status != http.StatusBadRequest {
				t.Fatalf("status %d: %s", status, response)
			}
			expectedCode := "invalid_configuration"
			if test.name == "unreadable active credential" {
				expectedCode = "credentials_unreadable"
			}
			if !strings.Contains(response, `"code":"`+expectedCode+`"`) {
				t.Fatalf("local failure lost %s classification: %s", expectedCode, response)
			}
			if f.runtime.calls.Load() != 0 || f.providerRequests.Load() != 0 {
				t.Fatal("invalid candidate reached runtime/provider")
			}
			if bytes.Contains([]byte(response), []byte("private-secret")) || bytes.Contains([]byte(response), []byte("secret-invalid-ca")) {
				t.Fatalf("private config leaked in public error: %s", response)
			}
		})
	}
}

func TestSessionValidateBeforeCreateAndStoredCredentialReuse(t *testing.T) {
	t.Run("before create", func(t *testing.T) {
		f := newSessionValidateFixture(t)
		body := f.body()
		delete(body, "instanceId")
		body["authContext"] = "dashboard"
		body["auth"] = map[string]any{"email": "new@example.test", "password": "new-password"}
		status, response := f.request(t, body)
		if status != http.StatusOK {
			t.Fatalf("status %d: %s", status, response)
		}
		if f.runtime.input.Instance.ID != "" || f.runtime.input.Auth["password"] != "new-password" {
			t.Fatal("new instance validation used stored credentials")
		}
	})
	t.Run("blank secret keeps stored value", func(t *testing.T) {
		f := newSessionValidateFixture(t)
		auth := pluginsystem.JSONMapFromRecord(f.instance, "auth")
		encrypted, err := security.Encrypt([]byte("saved-password"), oauthMetadataTestKey)
		if err != nil {
			t.Fatal(err)
		}
		auth["password"] = encrypted
		f.instance.Set("auth", auth)
		if err := f.app.Save(f.instance); err != nil {
			t.Fatal(err)
		}
		body := f.body()
		body["auth"] = map[string]any{"password": ""}
		status, response := f.request(t, body)
		if status != http.StatusOK {
			t.Fatalf("status %d: %s", status, response)
		}
		if f.runtime.input.Auth["password"] != "saved-password" {
			t.Fatal("blank secret did not reuse stored value")
		}
	})
}

func TestSessionValidateKeepsExistingNetworkRestriction(t *testing.T) {
	f := newSessionValidateFixture(t)
	body := f.body()
	body["config"].(map[string]any)["host"].(map[string]any)["connectors"].(map[string]any)["server"].(map[string]any)["allowPrivate"] = false
	status, response := f.request(t, body)
	if status != http.StatusBadRequest {
		t.Fatalf("status %d: %s", status, response)
	}
	if f.runtime.calls.Load() != 1 || f.providerRequests.Load() != 0 {
		t.Fatal("private provider received a request without explicit allowance")
	}
}

func TestSessionValidateDraftKeepsTypedGuestFailureWithoutPersisting(t *testing.T) {
	for _, test := range []struct {
		status int
		code   string
		retry  int
	}{{http.StatusTooManyRequests, "rate_limited", 0}, {http.StatusServiceUnavailable, "provider_unavailable", 120}} {
		t.Run(test.code, func(t *testing.T) {
			f := newSessionValidateFixture(t)
			originalAuth, originalConfig := f.instance.GetString("auth"), f.instance.GetString("config")
			f.runtime.guestError = pluginsystem.PluginCallError{PluginID: f.manifest.ID, Export: "refresh_auth_v1", PluginError: pluginsystem.PluginError{Code: test.code, Message: "provider-body-secret replacement-password stored@example.test " + f.baseURL + "?token=token-url-secret", Operation: pluginsystem.OperationLogin, HTTPStatus: test.status, RetryAfterSeconds: &test.retry}}
			status, response := f.request(t, f.body())
			if status != http.StatusBadRequest {
				t.Fatalf("typed guest failure status %d: %s", status, response)
			}
			var payload struct {
				Data struct {
					Operation struct {
						Code   string
						Params map[string]any
					}
				}
			}
			if err := json.Unmarshal([]byte(response), &payload); err != nil {
				t.Fatal(err)
			}
			operation := payload.Data.Operation
			if operation.Code != test.code || operation.Params["operation"] != pluginsystem.OperationLogin || operation.Params["httpStatus"] != float64(test.status) || operation.Params["retryAfterSeconds"] != float64(test.retry) {
				t.Fatalf("typed guest metadata lost: %s", response)
			}
			if f.runtime.calls.Load() != 1 || f.providerRequests.Load() != 1 {
				t.Fatal("replacement/draft did not reach its provider exactly once")
			}
			if f.runtime.input.Auth["password"] != "replacement-password" || f.runtime.input.Auth["email"] != "stored@example.test" || len(f.runtime.input.Auth) != 2 {
				t.Fatalf("typed failure used wrong active credentials: %#v", f.runtime.input.Auth)
			}
			if f.runtime.input.Config["limit"] != float64(3) || f.runtime.input.Config["after"] != "2026-01-01" {
				t.Fatalf("typed failure used wrong effective plugin config: %#v", f.runtime.input.Config)
			}
			connector := f.runtime.policy.Connectors["api"]
			if connector.BaseURL != f.baseURL || connector.BasePath != "/draft" || !connector.AllowPrivate {
				t.Fatalf("typed failure used old connector policy: %#v", connector)
			}
			for _, secret := range []string{"provider-body-secret", "replacement-password", "stored@example.test", "token-url-secret", f.baseURL} {
				if bytes.Contains([]byte(response), []byte(secret)) {
					t.Fatalf("guest failure leaked %q: %s", secret, response)
				}
			}
			fresh, err := f.app.FindRecordById("plugin_instances", f.instance.Id)
			if err != nil {
				t.Fatal(err)
			}
			if fresh.GetString("auth") != originalAuth || fresh.GetString("config") != originalConfig {
				t.Fatal("failed validation persisted draft credentials/config")
			}
		})
	}
}
