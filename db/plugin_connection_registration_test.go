package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Exercise production route/hook registration, rather than a fixture which
// redeclares the handlers and remains green when main.go omits a binding.
func TestPluginConnectionProductionRegistration(t *testing.T) {
	app := pocketbase.NewWithConfig(pocketbase.Config{DefaultDataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	instances := core.NewBaseCollection("plugin_instances")
	instances.CreateRule, instances.UpdateRule = types.Pointer(""), types.Pointer("")
	for _, name := range []string{"user", "plugin_id", "auth_context", "connection_generation", "oauth_token_revision", "status"} {
		instances.Fields.Add(&core.TextField{Name: name})
	}
	instances.Fields.Add(&core.JSONField{Name: "auth"}, &core.JSONField{Name: "config"}, &core.BoolField{Name: "enabled"})
	if err := app.Save(instances); err != nil {
		t.Fatal(err)
	}
	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatal(err)
	}
	superuser := core.NewRecord(superusers)
	superuser.SetEmail("registration@example.test")
	superuser.SetPassword("registration-test-password")
	if err := app.Save(superuser); err != nil {
		t.Fatal(err)
	}
	token, err := superuser.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	instance := core.NewRecord(instances)
	instance.Set("plugin_id", "registration-test")
	instance.Set("user", superuser.Id)
	instance.Set("connection_generation", "public-generation")
	instance.Set("oauth_token_revision", "private-token-revision")
	instance.Set("auth", map[string]any{"oauthState": "private-flow-state", "oauthCodeVerifier": "private-flow-verifier"})
	if err := app.Save(instance); err != nil {
		t.Fatal(err)
	}
	setupEventHandlers(app, nil)
	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	registerRoutes(&core.ServeEvent{App: app, Router: router}, nil)
	handler, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	request := func(method, path, body, auth string) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", auth)
		response, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, string(data)
	}
	for _, route := range []struct {
		method, path string
		status       int
	}{
		{http.MethodPost, "/plugins/instances", http.StatusForbidden},
		{http.MethodPatch, "/plugins/instances/" + instance.Id + "/settings", http.StatusForbidden},
		{http.MethodPost, "/plugins/oauth/start", http.StatusUnauthorized},
		{http.MethodPost, "/plugins/oauth/callback", http.StatusUnauthorized},
		{http.MethodPost, "/plugins/oauth/revoke", http.StatusUnauthorized},
		{http.MethodPost, "/plugins/trail-send", http.StatusUnauthorized},
		{http.MethodPost, "/plugins/auth/validate", http.StatusUnauthorized},
	} {
		if status, body := request(route.method, route.path, "{}", ""); status != route.status {
			t.Errorf("production %s %s returned %d: %s", route.method, route.path, status, body)
		}
	}
	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/api/collections/plugin_instances/records"},
		{http.MethodPatch, "/api/collections/plugin_instances/records/" + instance.Id},
	} {
		if status, body := request(route.method, route.path, `{ "enabled": false }`, token); status != http.StatusForbidden {
			t.Errorf("production direct write %s returned %d: %s", route.path, status, body)
		}
	}
	status, body := request(http.MethodGet, "/api/collections/plugin_instances/records/"+instance.Id, "", token)
	if status != http.StatusOK || !strings.Contains(body, "public-generation") {
		t.Fatalf("production read returned %d: %s", status, body)
	}
	for _, secret := range []string{"oauth_token_revision", "private-token-revision", "oauthState", "private-flow-state", "oauthCodeVerifier", "private-flow-verifier"} {
		if strings.Contains(body, secret) {
			t.Errorf("production Enrich binding leaked %s: %s", secret, body)
		}
	}
}
