package routes

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"testing"

	"github.com/pocketbase/pocketbase/tools/security"
	"pocketbase/pluginsystem"
	"pocketbase/util"
)

const legacyStravaTestSecret = "0123456789abcdef0123456789abcdef01234567"
const legacyStravaTestRefresh = "abcdef0123456789abcdef0123456789abcdef01"
const legacyStravaTestAccess = "fedcba9876543210fedcba9876543210fedcba98"

func rawLegacyStravaCredentials(t *testing.T, f *oauthMetadataFixture) {
	t.Helper()
	// A real pre-plugin migration wrote these raw values without model hooks.
	// They decode as 30 Base64 bytes and thus resemble unmarked PB ciphertext.
	decoded, err := base64.StdEncoding.DecodeString(legacyStravaTestSecret)
	if err != nil || len(decoded) != 30 || !util.LooksLikeEncrypted(legacyStravaTestSecret) {
		t.Fatal("fixture does not model ambiguous 40-hex credentials")
	}
	auth := rawOAuthMetadataAuth(t, f)
	auth["clientId"] = legacyStravaTestAccess // non-secret, despite the same shape
	auth["clientSecret"] = legacyStravaTestSecret
	auth["accessToken"] = legacyStravaTestAccess
	auth["refreshToken"] = legacyStravaTestRefresh
	replaceRawOAuthMetadataAuth(t, f, auth)
	f.expectedClientID = legacyStravaTestAccess
	f.expectedClientSecret = legacyStravaTestSecret
	f.expectedRefreshToken = legacyStravaTestRefresh
}

func assertStoredSecretEncryptedAs(t *testing.T, f *oauthMetadataFixture, field, expected string) {
	t.Helper()
	stored, ok := rawOAuthMetadataAuth(t, f)[field].(string)
	if !ok || stored == expected {
		t.Fatalf("stored %s remained raw plaintext", field)
	}
	plain, err := security.Decrypt(stored, oauthMetadataTestKey)
	if err != nil || string(plain) != expected {
		t.Fatalf("stored %s was not encrypted once with the current key", field)
	}
}

func TestRawLegacyOAuthSettingsDistinguishesInputFromPreservedCredentials(t *testing.T) {
	for _, secret := range []string{"short-legacy-secret", legacyStravaTestSecret} {
		for _, submission := range []string{"omitted", "empty", "unchanged"} {
			t.Run(secret+"/"+submission, func(t *testing.T) {
				f := newOAuthMetadataFixture(t)
				auth := rawOAuthMetadataAuth(t, f)
				auth["clientSecret"] = secret
				replaceRawOAuthMetadataAuth(t, f, auth)
				patch := map[string]any{"clientId": "strava-test-client"}
				switch submission {
				case "empty":
					patch["clientSecret"] = ""
				case "unchanged":
					patch["clientSecret"] = secret
				}
				f.saveModal(t, map[string]any{"auth": patch})
				if submission == "unchanged" {
					assertStoredSecretEncryptedAs(t, f, "clientSecret", secret)
				} else if rawOAuthMetadataAuth(t, f)["clientSecret"] != secret {
					t.Fatal("omitted/placeholder input modified retained credential bytes")
				}
			})
		}
	}
}

func TestRawLegacyStravaRefreshEncryptsOnlyItsNewTokens(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	rawLegacyStravaCredentials(t, f)
	instance, err := f.app.FindRecordById("plugin_instances", f.instanceID)
	if err != nil {
		t.Fatal(err)
	}
	auth, err := decryptedActiveInstanceAuth(f.plugin, instance)
	if err != nil {
		t.Fatalf("raw migrated credentials were treated as broken ciphertext: %v", err)
	}
	for field, expected := range map[string]string{"clientId": legacyStravaTestAccess, "clientSecret": legacyStravaTestSecret, "accessToken": legacyStravaTestAccess, "refreshToken": legacyStravaTestRefresh} {
		if auth[field] != expected {
			t.Fatalf("raw migrated %s changed before provider request", field)
		}
	}
	if _, err := pluginsystem.RefreshOAuthAuthIfNeeded(context.Background(), f.app, f.plugin, instance, auth); err != nil {
		t.Fatal(err)
	}
	if f.tokenRequests.Load() != 1 {
		t.Fatal("refresh did not make exactly one provider request")
	}
	if rawOAuthMetadataAuth(t, f)["clientSecret"] != legacyStravaTestSecret {
		t.Fatal("token/flow writer changed retained client credential bytes")
	}
	assertStoredSecretEncryptedAs(t, f, "accessToken", "strava-test-access-rotated")
	assertStoredSecretEncryptedAs(t, f, "refreshToken", "strava-test-refresh-rotated")
	if rawOAuthMetadataAuth(t, f)["clientId"] != legacyStravaTestAccess {
		t.Fatal("non-secret clientId was modified")
	}
}

func TestRawLegacyStravaReconnectAndRevokePreserveClientCredentials(t *testing.T) {
	for _, operation := range []string{"reconnect", "revoke"} {
		t.Run(operation, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			rawLegacyStravaCredentials(t, f)
			if operation == "revoke" {
				f.request(t, http.MethodPost, "/api/plugin-system/oauth/revoke", map[string]any{"instanceId": f.instanceID, "expected_connection_generation": oauthFixtureGeneration(t, f)})
				if f.tokenRequests.Load() != 0 {
					t.Fatal("revoke called provider")
				}
				f.assertAuthFieldsAbsent(t, "raw legacy revoke", "accessToken", "refreshToken")
			} else {
				start := f.request(t, http.MethodPost, "/api/plugin-system/oauth/start", map[string]any{"pluginId": "strava", "instanceId": f.instanceID, "expected_connection_generation": oauthFixtureGeneration(t, f), "redirectUri": "https://wanderer.example/settings/plugins/oauth/callback"})
				if f.tokenRequests.Load() != 0 {
					t.Fatal("start called token provider")
				}
				stored := rawOAuthMetadataAuth(t, f)
				if stored["accessToken"] != legacyStravaTestAccess || stored["refreshToken"] != legacyStravaTestRefresh {
					t.Fatal("OAuth start modified the existing grant")
				}
				f.request(t, http.MethodPost, "/api/plugin-system/oauth/callback", map[string]any{"instanceId": f.instanceID, "state": start["state"], "code": "test-authorization-code"})
				if f.tokenRequests.Load() != 1 {
					t.Fatal("callback did not exchange one code")
				}
				assertStoredSecretEncryptedAs(t, f, "accessToken", "strava-test-access-rotated")
			}
			if rawOAuthMetadataAuth(t, f)["clientSecret"] != legacyStravaTestSecret {
				t.Fatal("token/flow writer changed retained client credential bytes")
			}
			if rawOAuthMetadataAuth(t, f)["clientId"] != legacyStravaTestAccess {
				t.Fatal("non-secret clientId was modified")
			}
		})
	}
}

func TestActiveAuthUsesWriterUnionForSharedInputFields(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	manifest := f.plugin.Manifest
	active := manifest.Auth.Contexts["oauth_access_token"]
	active.Fields = append(active.Fields, "shared")
	manifest.Auth.Contexts["oauth_access_token"] = active
	manifest.Auth.Contexts["inactive_api"] = pluginsystem.AuthContext{Type: pluginsystem.AuthTypeAPIKey, SecretField: "shared"}
	manifest.Permissions.Auth = append(manifest.Permissions.Auth, "inactive_api")
	if err := pluginsystem.ValidateManifest(manifest); err != nil {
		t.Fatalf("valid shared-field manifest rejected: %v", err)
	}
	installed, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", "strava")
	if err != nil {
		t.Fatal(err)
	}
	installed.Set("manifest", manifest)
	if err := f.app.Save(installed); err != nil {
		t.Fatal(err)
	}
	f.plugin.Manifest = manifest
	instance, _ := f.storedAuth(t)
	if err := pluginsystem.ApplyAuthPatch(instance, manifest, map[string]any{"shared": "shared-input-secret"}); err != nil {
		t.Fatal(err)
	}
	if err := f.app.Save(instance); err != nil {
		t.Fatal(err)
	}
	assertStoredSecretEncryptedAs(t, f, "shared", "shared-input-secret")
	instance, err = f.app.FindRecordById("plugin_instances", f.instanceID)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := decryptedActiveInstanceAuth(f.plugin, instance)
	if err != nil {
		t.Fatal(err)
	}
	guest := pluginsystem.PluginInputAuth(f.plugin, decrypted, instance.GetString("auth_context"))
	if guest["shared"] != "shared-input-secret" {
		t.Fatal("active guest received the ciphertext of a shared secret field")
	}
	flow, err := decryptedOAuthFlowAuth(instance, manifest, active, false)
	if err != nil || flow["shared"] != "shared-input-secret" {
		t.Fatal("OAuth flow did not use the writer's shared field definition")
	}
}

func TestActiveAuthDoesNotTreatUnrelatedCiphertextAsLegacyPlaintext(t *testing.T) {
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", oauthMetadataTestKey)
	foreign, err := security.Encrypt([]byte("foreign-secret"), "ffffffffffffffffffffffffffffffff")
	if err != nil {
		t.Fatal(err)
	}
	corruptBytes, err := base64.StdEncoding.DecodeString(foreign)
	if err != nil {
		t.Fatal(err)
	}
	corruptBytes[len(corruptBytes)-1] ^= 1
	corrupted := base64.StdEncoding.EncodeToString(corruptBytes)
	for _, tc := range []struct{ name, pluginID, field, value string }{
		{"foreign key", "strava", "clientSecret", foreign},
		{"corrupted", "strava", "clientSecret", corrupted},
		{"other provider hex", "hammerhead", "clientSecret", legacyStravaTestSecret},
		{"other secret field hex", "strava", "password", legacyStravaTestSecret},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeSession, Fields: []string{tc.field}, SecretFields: []string{tc.field}}
			manifest := pluginsystem.Manifest{ID: tc.pluginID, Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{"active": ctx}}}
			auth := map[string]any{tc.field: tc.value}
			if err := decryptActiveAuth(auth, manifest, ctx, nil); !errors.Is(err, pluginsystem.ErrActiveCredentialsUnreadable) {
				t.Fatalf("unreadable secret was accepted: %v", err)
			}
			if auth[tc.field] != tc.value {
				t.Fatal("unreadable ciphertext was changed")
			}
		})
	}
}

func TestActiveAuthDecryptsStoredLegacyFormatSecret(t *testing.T) {
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", oauthMetadataTestKey)
	encrypted, err := security.Encrypt([]byte(legacyStravaTestSecret), oauthMetadataTestKey)
	if err != nil {
		t.Fatal(err)
	}
	ctx := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeOAuth2}
	manifest := pluginsystem.Manifest{ID: "strava", Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{"oauth": ctx}}}
	auth := map[string]any{"clientSecret": encrypted}
	if err := decryptActiveAuth(auth, manifest, ctx, nil); err != nil || auth["clientSecret"] != legacyStravaTestSecret {
		t.Fatal("stored current-key ciphertext was not decrypted")
	}
}
