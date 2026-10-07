package routes

import (
	"errors"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"pocketbase/pluginsystem"
	"testing"
)

func TestActiveAuthIgnoresUnreadableInactiveAndPendingSecrets(t *testing.T) {
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", oauthMetadataTestKey)
	broken, err := security.Encrypt([]byte("unreadable-secret"), "ffffffffffffffffffffffffffffffff")
	if err != nil {
		t.Fatal(err)
	}
	valid, err := security.Encrypt([]byte("active-password"), oauthMetadataTestKey)
	if err != nil {
		t.Fatal(err)
	}
	instance := core.NewRecord(core.NewBaseCollection("instances"))
	instance.Collection().Fields.Add(&core.JSONField{Name: "auth"}, &core.TextField{Name: "auth_context"})
	// A non-secret input can resemble encrypted data; the writer stores it
	// unchanged, so the reader must never attempt to decrypt it.
	legacyEmail := "0123456789abcdef0123456789abcdef01234567"
	instance.Set("auth_context", "web")
	instance.Set("auth", map[string]any{"oauthContext": "web", "email": legacyEmail, "password": valid, "refreshToken": broken, "oauthState": broken, "apiKey": broken})
	plugin := pluginsystem.LocalPlugin{Manifest: pluginsystem.Manifest{Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{"web": {Type: pluginsystem.AuthTypeSession, Fields: []string{"email", "password"}, SecretFields: []string{"password"}}, "api": {Type: pluginsystem.AuthTypeAPIKey, SecretField: "apiKey"}}}}}
	auth, err := decryptedActiveInstanceAuth(plugin, instance)
	if err != nil {
		t.Fatal(err)
	}
	if auth["password"] != "active-password" || auth["refreshToken"] != broken || auth["oauthState"] != broken {
		t.Fatalf("incorrect decryption/preservation: %#v", auth)
	}
	guest := pluginsystem.PluginInputAuth(plugin, auth, instance.GetString("auth_context"))
	if len(guest) != 2 || guest["email"] != legacyEmail || guest["password"] != "active-password" {
		t.Fatalf("inactive secrets entered guest input: %#v", guest)
	}
	auth["password"] = broken
	if err := decryptActiveAuth(auth, plugin.Manifest, plugin.Manifest.Auth.Contexts["web"], nil); !errors.Is(err, pluginsystem.ErrActiveCredentialsUnreadable) {
		t.Fatalf("active unreadable secret: %v", err)
	}
	auth["password"] = "new-password"
	if err := decryptActiveAuth(auth, plugin.Manifest, plugin.Manifest.Auth.Contexts["web"], map[string]bool{"password": true}); err != nil {
		t.Fatalf("replacement did not override unreadable stored value: %v", err)
	}
}

func TestOAuthFlowDecryptsOnlyItsOwnPrerequisites(t *testing.T) {
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", oauthMetadataTestKey)
	encrypt := func(value, key string) string {
		result, err := security.Encrypt([]byte(value), key)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	client := encrypt("client-secret", oauthMetadataTestKey)
	state := encrypt("pending-state", oauthMetadataTestKey)
	verifier := encrypt("pkce-verifier", oauthMetadataTestKey)
	unreadable := encrypt("obsolete-secret", "ffffffffffffffffffffffffffffffff")
	collection := core.NewBaseCollection("instances")
	collection.Fields.Add(&core.JSONField{Name: "auth"})
	record := core.NewRecord(collection)
	record.Set("auth", map[string]any{"clientSecret": client, "accessToken": unreadable, "refreshToken": unreadable, "password": unreadable, "oauthState": state, "oauthCodeVerifier": verifier})
	context := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeOAuth2, PKCE: true, SecretFields: []string{"clientSecret", "accessToken", "refreshToken"}}
	manifest := pluginsystem.Manifest{ID: "test", Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{"oauth": context}}}
	for _, callback := range []bool{false, true} {
		auth, err := decryptedOAuthFlowAuth(record, manifest, context, callback)
		if err != nil {
			t.Fatal(err)
		}
		if auth["clientSecret"] != "client-secret" {
			t.Fatal("flow did not decrypt its client credential")
		}
		for _, key := range []string{"accessToken", "refreshToken", "password"} {
			if auth[key] != unreadable {
				t.Fatalf("flow decrypted unrelated %s", key)
			}
		}
		if callback {
			if auth["oauthState"] != "pending-state" || auth["oauthCodeVerifier"] != "pkce-verifier" {
				t.Fatal("callback did not decrypt its flow")
			}
		} else if auth["oauthState"] != state || auth["oauthCodeVerifier"] != verifier {
			t.Fatal("start decrypted a previous flow")
		}
	}
}

func TestOAuthFlowWithoutPKCEDoesNotDecryptObsoleteVerifier(t *testing.T) {
	t.Setenv("POCKETBASE_ENCRYPTION_KEY", oauthMetadataTestKey)
	state, err := security.Encrypt([]byte("current-pending-state"), oauthMetadataTestKey)
	if err != nil {
		t.Fatal(err)
	}
	obsoleteVerifier, err := security.Encrypt([]byte("previous-pkce-verifier"), "ffffffffffffffffffffffffffffffff")
	if err != nil {
		t.Fatal(err)
	}
	collection := core.NewBaseCollection("instances")
	collection.Fields.Add(&core.JSONField{Name: "auth"})
	record := core.NewRecord(collection)
	record.Set("auth", map[string]any{"oauthState": state, "oauthCodeVerifier": obsoleteVerifier})
	context := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeOAuth2}
	manifest := pluginsystem.Manifest{ID: "test", Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{"oauth": context}}}
	auth, err := decryptedOAuthFlowAuth(record, manifest, context, true)
	if err != nil {
		t.Fatal("non-PKCE callback depended on an obsolete verifier")
	}
	if auth["oauthState"] != "current-pending-state" || auth["oauthCodeVerifier"] != obsoleteVerifier {
		t.Fatal("non-PKCE callback did not limit decryption to its current state")
	}
	context.PKCE = true
	if _, err := decryptedOAuthFlowAuth(record, manifest, context, true); !errors.Is(err, pluginsystem.ErrActiveCredentialsUnreadable) {
		t.Fatal("PKCE callback ignored its unreadable required verifier")
	}
}
