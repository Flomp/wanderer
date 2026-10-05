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
	instance.Collection().Fields.Add(&core.JSONField{Name: "auth"})
	legacyEmail, err := security.Encrypt([]byte("user@example.test"), oauthMetadataTestKey)
	if err != nil {
		t.Fatal(err)
	}
	instance.Set("auth", map[string]any{"oauthContext": "web", "email": legacyEmail, "password": valid, "refreshToken": broken, "oauthState": broken, "apiKey": broken})
	plugin := pluginsystem.LocalPlugin{Manifest: pluginsystem.Manifest{Auth: pluginsystem.AuthManifest{Contexts: map[string]pluginsystem.AuthContext{"web": {Type: pluginsystem.AuthTypeSession, Fields: []string{"email", "password"}, SecretFields: []string{"password"}}, "api": {Type: pluginsystem.AuthTypeAPIKey, SecretField: "apiKey"}}}}}
	auth, err := decryptedActiveInstanceAuth(plugin, instance)
	if err != nil {
		t.Fatal(err)
	}
	if auth["password"] != "active-password" || auth["refreshToken"] != broken || auth["oauthState"] != broken {
		t.Fatalf("incorrect decryption/preservation: %#v", auth)
	}
	guest := pluginsystem.PluginInputAuth(plugin, auth)
	if len(guest) != 2 || guest["email"] != "user@example.test" || guest["password"] != "active-password" {
		t.Fatalf("inactive secrets entered guest input: %#v", guest)
	}
	auth["password"] = broken
	if err := decryptActiveAuth(auth, plugin.Manifest.Auth.Contexts["web"], nil); !errors.Is(err, pluginsystem.ErrActiveCredentialsUnreadable) {
		t.Fatalf("active unreadable secret: %v", err)
	}
	auth["password"] = "new-password"
	if err := decryptActiveAuth(auth, plugin.Manifest.Auth.Contexts["web"], map[string]bool{"password": true}); err != nil {
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
	context := pluginsystem.AuthContext{Type: pluginsystem.AuthTypeOAuth2, SecretFields: []string{"clientSecret", "accessToken", "refreshToken"}}
	for _, callback := range []bool{false, true} {
		auth, err := decryptedOAuthFlowAuth(record, context, callback)
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
