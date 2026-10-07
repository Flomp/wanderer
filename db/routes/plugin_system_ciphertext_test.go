package routes

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/security"

	"pocketbase/pluginsystem"
)

func rawOAuthMetadataAuth(t *testing.T, f *oauthMetadataFixture) map[string]any {
	t.Helper()
	instance, err := f.app.FindRecordById("plugin_instances", f.instanceID)
	if err != nil {
		t.Fatal(err)
	}
	return pluginsystem.JSONMapFromRecord(instance, "auth")
}

func replaceRawOAuthMetadataAuth(t *testing.T, f *oauthMetadataFixture, auth map[string]any) {
	t.Helper()
	encoded, err := json.Marshal(auth)
	if err != nil {
		t.Fatal(err)
	}
	// Model stored data encrypted before a key change without submitting a new
	// credential through the normal create/update encryption hook.
	if _, err := f.app.DB().NewQuery("UPDATE plugin_instances SET auth={:auth} WHERE id={:id}").Bind(dbx.Params{"auth": string(encoded), "id": f.instanceID}).Execute(); err != nil {
		t.Fatal(err)
	}
}

func assertUnreadableOAuthSecretUnchanged(t *testing.T, f *oauthMetadataFixture, field, expected string) {
	t.Helper()
	stored, ok := rawOAuthMetadataAuth(t, f)[field].(string)
	if !ok || stored != expected {
		t.Fatalf("stored %s changed while preserving an unrelated credential", field)
	}
	if _, err := security.Decrypt(stored, os.Getenv("POCKETBASE_ENCRYPTION_KEY")); err == nil {
		t.Fatalf("unreadable %s became decryptable; its ciphertext must not be encrypted again", field)
	}
}

func TestOAuthMetadataRefreshPreservesUnreadableInactiveSecret(t *testing.T) {
	f := newOAuthMetadataFixture(t)
	installed, err := f.app.FindFirstRecordByData("installed_plugins", "plugin_id", "strava")
	if err != nil {
		t.Fatal(err)
	}
	manifest := f.plugin.Manifest
	manifest.Auth.Contexts["inactive_dashboard"] = pluginsystem.AuthContext{
		Type: pluginsystem.AuthTypeSession, Fields: []string{"email", "password"}, SecretFields: []string{"password"},
		Refresh: &pluginsystem.AuthRefresh{Mode: pluginsystem.AuthRefreshModePlugin, Function: "refresh_auth_v1"},
	}
	manifest.Permissions.Auth = append(manifest.Permissions.Auth, "inactive_dashboard")
	installed.Set("manifest", manifest)
	if err := f.app.Save(installed); err != nil {
		t.Fatal(err)
	}
	f.plugin.Manifest = manifest
	unreadable, err := security.Encrypt([]byte("inactive-dashboard-password"), "ffffffffffffffffffffffffffffffff")
	if err != nil {
		t.Fatal(err)
	}
	stored := rawOAuthMetadataAuth(t, f)
	stored["password"] = unreadable
	replaceRawOAuthMetadataAuth(t, f, stored)
	instance, err := f.app.FindRecordById("plugin_instances", f.instanceID)
	if err != nil {
		t.Fatal(err)
	}
	auth, err := decryptedActiveInstanceAuth(f.plugin, instance)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pluginsystem.RefreshOAuthAuthIfNeeded(context.Background(), f.app, f.plugin, instance, auth); err != nil {
		t.Fatal(err)
	}
	if f.tokenRequests.Load() != 1 {
		t.Fatal("refresh did not make exactly one token request")
	}
	assertUnreadableOAuthSecretUnchanged(t, f, "password", unreadable)
	_, saved := f.storedAuth(t)
	if saved["accessToken"] != "strava-test-access-rotated" || saved["refreshToken"] != "strava-test-refresh-rotated" {
		t.Fatal("refresh did not save the newly issued grant")
	}
}

func TestOAuthMetadataSettingsPreservesUnreadableSecretUntilExplicitReplacement(t *testing.T) {
	for _, submission := range []string{"omitted", "empty", "explicit ciphertext-shaped input", "replacement", "explicit null"} {
		t.Run(submission, func(t *testing.T) {
			f := newOAuthMetadataFixture(t)
			unreadable, err := security.Encrypt([]byte("previous-client-secret"), "ffffffffffffffffffffffffffffffff")
			if err != nil {
				t.Fatal(err)
			}
			stored := rawOAuthMetadataAuth(t, f)
			stored["clientSecret"] = unreadable
			replaceRawOAuthMetadataAuth(t, f, stored)
			auth := map[string]any{"clientId": "strava-test-client"}
			switch submission {
			case "empty":
				auth["clientSecret"] = ""
			case "explicit ciphertext-shaped input":
				auth["clientSecret"] = unreadable
			case "replacement":
				auth["clientSecret"] = "replacement-client-secret"
			case "explicit null":
				auth["clientSecret"] = nil
			}
			f.saveModal(t, map[string]any{"auth": auth})
			switch submission {
			case "replacement", "explicit ciphertext-shaped input":
				expected := "replacement-client-secret"
				if submission == "explicit ciphertext-shaped input" {
					expected = unreadable
				}
				ciphertext, ok := rawOAuthMetadataAuth(t, f)["clientSecret"].(string)
				if !ok || ciphertext == expected || ciphertext == unreadable {
					t.Fatal("explicit replacement was not newly encrypted")
				}
				plaintext, err := security.Decrypt(ciphertext, oauthMetadataTestKey)
				if err != nil || string(plaintext) != expected {
					t.Fatal("replacement secret was not encrypted with the current key")
				}
			case "explicit null":
				f.assertAuthFieldsAbsent(t, "explicit null", "clientSecret")
			default:
				assertUnreadableOAuthSecretUnchanged(t, f, "clientSecret", unreadable)
			}
		})
	}
}
