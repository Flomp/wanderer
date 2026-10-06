package pluginsystem

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestEffectiveAuthContextRequiresAnUnambiguousSelection(t *testing.T) {
	manifest := Manifest{Auth: AuthManifest{Contexts: map[string]AuthContext{"web": {Type: AuthTypeSession, Fields: []string{"email", "password"}, SecretFields: []string{"password"}}, "api": {Type: AuthTypeOAuth2}}}}
	if _, _, err := ResolveAuthContext(manifest, ""); err == nil {
		t.Fatal("ambiguous manifest silently selected a context")
	}
	if _, _, err := ResolveAuthContext(manifest, "unknown"); err == nil {
		t.Fatal("unknown explicit context accepted")
	}
	name, _, err := EffectiveAuthContext(manifest, map[string]any{"oauthContext": "web"})
	if err != nil || name != "web" {
		t.Fatalf("explicit selection: %s, %v", name, err)
	}
	if _, _, err := EffectiveAuthContext(manifest, map[string]any{"oauthContext": 42}); err == nil {
		t.Fatal("invalid context value accepted")
	}
	delete(manifest.Auth.Contexts, "api")
	name, _, err = ResolveAuthContext(manifest, "")
	if err != nil || name != "web" {
		t.Fatalf("legacy single context: %s, %v", name, err)
	}
	if name, _, err = ResolveAuthContext(Manifest{}, ""); err != nil || name != "" {
		t.Fatalf("no-auth manifest: %s, %v", name, err)
	}
}

func TestActiveContextDoesNotRefreshInactiveOAuth(t *testing.T) {
	plugin := LocalPlugin{Manifest: Manifest{Auth: AuthManifest{Contexts: map[string]AuthContext{"web": {Type: AuthTypeSession, Fields: []string{"email"}}, "api": {Type: AuthTypeOAuth2}}}}}
	auth := map[string]any{"oauthContext": "web", "email": "user@example.test", "refreshToken": "inactive-refresh", "expiresAt": "1970-01-01T00:00:00Z"}
	got, err := RefreshOAuthAuthIfNeeded(context.Background(), nil, plugin, nil, auth)
	if err != nil || got["refreshToken"] != "inactive-refresh" {
		t.Fatalf("inactive refresh changed: %#v, %v", got, err)
	}
	guest := PluginInputAuth(plugin, auth)
	if len(guest) != 1 || guest["email"] != "user@example.test" {
		t.Fatalf("inactive credentials leaked: %#v", guest)
	}
}

func TestShippedStravaPluginInputRetainsAccessToken(t *testing.T) {
	data, err := os.ReadFile("../../plugins/strava/plugin.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	got := PluginInputAuth(LocalPlugin{Manifest: manifest}, map[string]any{"accessToken": "required-by-strava", "refreshToken": "host-only", "clientSecret": "host-secret", "oauthState": "pending-state", "email": "inactive-user", "expiresAt": "2030-01-01T00:00:00Z"})
	if got["accessToken"] != "required-by-strava" {
		t.Fatal("Strava accessToken removed")
	}
	for _, field := range []string{"refreshToken", "clientSecret", "oauthState", "email"} {
		if _, ok := got[field]; ok {
			t.Fatalf("unexpected guest field %s", field)
		}
	}
}
