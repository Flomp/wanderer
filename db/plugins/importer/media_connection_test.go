package importer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"pocketbase/pluginsystem"
)

func TestConnectorMediaRejectsInactiveAuthContextBeforeTransport(t *testing.T) {
	var requests atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Header.Get("Authorization") != "Bearer active-secret" {
			t.Error("media transport received an inactive credential")
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("image"))
	}))
	defer provider.Close()
	manifest := pluginsystem.Manifest{
		Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{
			"active":   {Type: pluginsystem.AuthTypeBearer, SecretField: "activeToken", SecretFields: []string{"activeToken"}},
			"inactive": {Type: pluginsystem.AuthTypeBearer, SecretField: "inactiveToken", SecretFields: []string{"inactiveToken"}},
		}},
		Permissions: pluginsystem.PermissionManifest{Auth: []string{"active", "inactive"}, Network: pluginsystem.NetworkPermissions{Connectors: []pluginsystem.ConnectorTargetPermission{
			{Name: "api", Type: pluginsystem.ConnectorTypeConfigured, ConfigKey: "server", AllowedPathPrefixes: []string{"/"}, SupportsMediaAuth: true, Auth: []string{"active", "inactive"}},
		}}},
	}
	auth := map[string]any{"activeToken": "active-secret", "inactiveToken": "inactive-secret"}
	policy := pluginsystem.RequestPolicyContext{Connectors: map[string]pluginsystem.ResolvedConnectorTarget{
		"api": {Name: "api", Type: pluginsystem.ConnectorTypeConfigured, BaseURL: provider.URL, BasePath: "/", AllowPrivate: true, AllowedPathPrefixes: []string{"/"}},
	}}
	opts := Options{Manifest: manifest, Auth: auth, Policy: policy.WithHostAuth(auth, "active")}
	photo := pluginsystem.Photo{Source: pluginsystem.MediaSource{MediaRef: &pluginsystem.MediaRef{Auth: "inactive", Connector: "api", Path: "/photo"}}}
	result, err := fetchConnectorMedia(context.Background(), photo, opts, 1024)
	if result != nil || err == nil || pluginsystem.PublicPluginError(err).Code != "invalid_configuration" || requests.Load() != 0 {
		t.Fatal("inactive media context was not rejected before transport")
	}
	// The same fully configured target and active credential pass the context
	// guard. Production then blocks literal loopback at its network boundary.
	photo.Source.MediaRef.Auth = "active"
	_, err = fetchConnectorMedia(context.Background(), photo, opts, 1024)
	if err == nil || pluginsystem.PublicPluginError(err).Code == "invalid_configuration" || requests.Load() != 0 {
		t.Fatal("active control did not reach the productive connector transport boundary")
	}
}
