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

func useFailingTokenProvider(t *testing.T, f *oauthMetadataFixture, status int, response string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	manifest := f.plugin.Manifest
	context := manifest.Auth.Contexts["oauth_access_token"]
	context.TokenURL = server.URL + "/oauth/token?token=token-url-secret"
	manifest.Auth.Contexts["oauth_access_token"] = context
	for i := range manifest.Permissions.Network.Connectors {
		if manifest.Permissions.Network.Connectors[i].Name == "oauth" {
			manifest.Permissions.Network.Connectors[i].FixedBaseURL = server.URL + "/oauth"
		}
	}
	f.plugin.Manifest = manifest
	record, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", manifest.ID)
	if err != nil {
		t.Fatal(err)
	}
	record.Set("manifest", manifest)
	if err := f.app.Save(record); err != nil {
		t.Fatal(err)
	}
	return server
}

func TestCronRefreshFailuresClassifyWithoutLeaking(t *testing.T) {
	for _, tc := range []struct {
		status            int
		body, code, state string
		network           bool
	}{
		{400, `{"error":"invalid_grant","error_description":"provider-body-secret"}`, "invalid_grant", "needs_reauth", false},
		{400, `{"error":"invalid_scope","error_description":"provider-body-secret"}`, "invalid_scope", "error", false},
		{401, `provider-body-secret`, "invalid_client", "needs_reauth", false},
		{403, `provider-body-secret`, "invalid_grant", "needs_reauth", false},
		{429, `provider-body-secret`, "rate_limited", "rate_limited", false},
		{503, `provider-body-secret`, "provider_unavailable", "unavailable", false},
		{503, `provider-body-secret`, "provider_unavailable", "unavailable", true},
	} {
		t.Run(tc.code+"/"+tc.state+"/"+http.StatusText(tc.status), func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			server := useFailingTokenProvider(t, f, tc.status, tc.body)
			if tc.network {
				server.Close()
			}
			collection, err := f.app.FindCollectionByNameOrId("plugin_instances")
			if err != nil {
				t.Fatal(err)
			}
			collection.Fields.Add(&core.DateField{Name: "retry_not_before"})
			if err := f.app.Save(collection); err != nil {
				t.Fatal(err)
			}
			actors := core.NewBaseCollection("activitypub_actors")
			actors.Fields.Add(&core.TextField{Name: "user"})
			if err := f.app.Save(actors); err != nil {
				t.Fatal(err)
			}
			instance, _ := f.storedAuth(t)
			actor := core.NewRecord(actors)
			actor.Set("user", instance.GetString("user"))
			if err := f.app.Save(actor); err != nil {
				t.Fatal(err)
			}
			_, err = syncPluginInstance(context.Background(), f.app, nil, f.plugin, instance)
			if err == nil {
				t.Fatal("expected refresh failure before runtime")
			}
			instance, _ = f.storedAuth(t)
			if state := instance.GetString("status"); state != tc.state {
				t.Fatalf("state %s, want %s", state, tc.state)
			}
			diagnostic := pluginsystem.JSONMapFromRecord(instance, "last_error")
			if diagnostic["code"] != tc.code || diagnostic["operation"] != pluginsystem.OperationOAuthRefresh {
				t.Fatalf("diagnostic %#v", diagnostic)
			}
			if tc.status == 429 && !tc.network {
				deadline := instance.GetDateTime("retry_not_before").Time()
				if deadline.Before(time.Now().Add(110*time.Second)) || deadline.After(time.Now().Add(130*time.Second)) {
					t.Fatalf("wrong retry deadline %v", deadline)
				}
			}
			encoded, _ := json.Marshal(diagnostic)
			for _, value := range []string{"provider-body-secret", "token-url-secret", "strava-test-client-secret", "strava-test-refresh", server.URL} {
				if strings.Contains(string(encoded), value) || strings.Contains(err.Error(), value) {
					t.Fatalf("provider leak %q in %s / %v", value, encoded, err)
				}
			}
		})
	}
}

func TestOAuthCallbackFailureKeepsSafeStructuredCode(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	useFailingTokenProvider(t, f, 400, `{"error":"invalid_scope","error_description":"provider-body-secret"}`)
	start := f.request(t, http.MethodPost, "/api/plugin-system/oauth/start", map[string]any{
		"pluginId": "strava", "instanceId": f.instanceID, "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback",
	})
	body, _ := json.Marshal(map[string]any{"instanceId": f.instanceID, "state": start["state"], "code": "test-authorization-code"})
	req := httptest.NewRequest(http.MethodPost, "/api/plugin-system/oauth/callback", bytes.NewReader(body))
	req.Header.Set("Authorization", f.authToken)
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, req)
	if response.Code != 400 {
		t.Fatalf("unexpected response %d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data struct {
			Operation struct {
				Code   string
				Params map[string]any
			}
		}
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.Operation.Code != "invalid_scope" || payload.Data.Operation.Params["operation"] != pluginsystem.OperationOAuthExchange || payload.Data.Operation.Params["httpStatus"] != float64(400) {
		t.Fatalf("lost error metadata: %s", response.Body.String())
	}
	for _, value := range []string{"provider-body-secret", "token-url-secret", "strava-test-client-secret"} {
		if strings.Contains(response.Body.String(), value) {
			t.Fatalf("callback leaked %q: %s", value, response.Body.String())
		}
	}
}

func TestPublicPluginOperationPayloadPreservesCodes(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	router, err := apis.NewRouter(f.app)
	if err != nil {
		t.Fatal(err)
	}
	router.GET("/safe-send-failure", func(e *core.RequestEvent) error {
		return pluginOperationAPIError(pluginsystem.PluginFailure{PluginError: pluginsystem.ClassifyProviderError(pluginsystem.OperationSend, 429, "0", "", time.Now())})
	})
	handler, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/safe-send-failure", nil))
	if response.Code != 400 || !strings.Contains(response.Body.String(), `"code":"rate_limited"`) || !strings.Contains(response.Body.String(), `"retryAfterSeconds":0`) {
		t.Fatalf("PocketBase lost public code/zero: %d %s", response.Code, response.Body.String())
	}
}

func TestCronInvalidStoredAuthContextIsConfigurationFailure(t *testing.T) {
	f := newOAuthMetadataFixture(t, func(auth map[string]any) { auth["oauthContext"] = "context-no-longer-declared" })
	collection, err := f.app.FindCollectionByNameOrId("plugin_instances")
	if err != nil {
		t.Fatal(err)
	}
	collection.Fields.Add(&core.DateField{Name: "retry_not_before"})
	if err := f.app.Save(collection); err != nil {
		t.Fatal(err)
	}
	actors := core.NewBaseCollection("activitypub_actors")
	actors.Fields.Add(&core.TextField{Name: "user"})
	if err := f.app.Save(actors); err != nil {
		t.Fatal(err)
	}
	instance, _ := f.storedAuth(t)
	actor := core.NewRecord(actors)
	actor.Set("user", instance.GetString("user"))
	if err := f.app.Save(actor); err != nil {
		t.Fatal(err)
	}
	_, err = syncPluginInstance(context.Background(), f.app, nil, f.plugin, instance)
	if err == nil {
		t.Fatal("Cron accepted an undeclared stored auth context")
	}
	if safe := pluginsystem.PublicPluginError(err); safe.Code != "invalid_request" {
		t.Fatalf("configuration failure classified as %#v", safe)
	}
	fresh, findErr := f.app.FindRecordById("plugin_instances", f.instanceID)
	if findErr != nil {
		t.Fatal(findErr)
	}
	if fresh.GetString("status") != "error" {
		t.Fatalf("invalid context status = %q, want error", fresh.GetString("status"))
	}
	diagnostic := pluginsystem.JSONMapFromRecord(fresh, "last_error")
	if diagnostic["code"] != "invalid_request" || diagnostic["operation"] != pluginsystem.OperationLogin {
		t.Fatalf("stored invalid-context diagnostic = %#v", diagnostic)
	}
	if f.tokenRequests.Load() != 0 {
		t.Fatal("invalid context reached the provider")
	}
}
