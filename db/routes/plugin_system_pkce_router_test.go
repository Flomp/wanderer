package routes

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"
	"testing"

	"github.com/pocketbase/pocketbase/tools/security"
)

func TestOAuthRouterPKCEStartAndCallbackUseCurrentVerifier(t *testing.T) {
	for _, pkce := range []bool{false, true} {
		name := "non-PKCE"
		if pkce {
			name = "PKCE"
		}
		t.Run(name, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			manifest := f.plugin.Manifest
			authContext := manifest.Auth.Contexts["oauth_access_token"]
			authContext.PKCE = pkce
			manifest.Auth.Contexts["oauth_access_token"] = authContext
			installed, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", "strava")
			if err != nil {
				t.Fatal(err)
			}
			installed.Set("manifest", manifest)
			if err := f.app.Save(installed); err != nil {
				t.Fatal(err)
			}
			f.plugin.Manifest = manifest
			auth := rawOAuthMetadataAuth(t, f)
			obsoleteVerifier, err := security.Encrypt([]byte("obsolete-verifier"), "ffffffffffffffffffffffffffffffff")
			if err != nil {
				t.Fatal(err)
			}
			auth["oauthCodeVerifier"] = obsoleteVerifier
			replaceRawOAuthMetadataAuth(t, f, auth)
			start := f.request(t, http.MethodPost, "/api/plugin-system/oauth/start", map[string]any{"pluginId": "strava", "instanceId": f.instanceID, "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback"})
			if f.tokenRequests.Load() != 0 {
				t.Fatal("OAuth start contacted token provider")
			}
			authorizationURL, err := url.Parse(start["url"].(string))
			if err != nil {
				t.Fatal(err)
			}
			stored := rawOAuthMetadataAuth(t, f)
			if pkce {
				encrypted, ok := stored["oauthCodeVerifier"].(string)
				if !ok || encrypted == obsoleteVerifier {
					t.Fatal("PKCE start did not replace its verifier")
				}
				verifier, err := security.Decrypt(encrypted, oauthMetadataTestKey)
				if err != nil || len(verifier) < 43 {
					t.Fatal("PKCE verifier was not stored encrypted")
				}
				f.expectedPKCEVerifier = string(verifier)
				digest := sha256.Sum256(verifier)
				if authorizationURL.Query().Get("code_challenge_method") != "S256" || authorizationURL.Query().Get("code_challenge") != base64.RawURLEncoding.EncodeToString(digest[:]) {
					t.Fatal("authorization challenge does not match the saved verifier")
				}
			} else {
				if _, exists := stored["oauthCodeVerifier"]; exists {
					t.Fatal("non-PKCE start kept an obsolete verifier")
				}
				if authorizationURL.Query().Has("code_challenge") || authorizationURL.Query().Has("code_challenge_method") {
					t.Fatal("non-PKCE start sent a challenge")
				}
				// Even if an obsolete field is present at callback time, a
				// non-PKCE exchange must neither decrypt it nor send it.
				stored["oauthCodeVerifier"] = obsoleteVerifier
				replaceRawOAuthMetadataAuth(t, f, stored)
			}
			f.request(t, http.MethodPost, "/api/plugin-system/oauth/callback", map[string]any{"instanceId": f.instanceID, "state": start["state"], "code": "test-authorization-code"})
			if f.tokenRequests.Load() != 1 {
				t.Fatal("callback did not make exactly one token request")
			}
			f.assertAuthFieldsAbsent(t, "completed PKCE flow", "oauthState", "oauthCodeVerifier", "oauthRedirectURI")
		})
	}
}
