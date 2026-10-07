package pluginsystem

import "testing"

func TestManifestAuthSecretFieldsIncludesAllContexts(t *testing.T) {
	manifest := Manifest{Auth: AuthManifest{Contexts: map[string]AuthContext{
		"active":   {Fields: []string{"email", "shared"}, SecretFields: []string{"password"}},
		"inactive": {SecretField: "shared", SecretFields: []string{"apiKey"}},
	}}}
	fields := ManifestAuthSecretFields(manifest)
	for _, field := range append(InternalAuthSecretFields(), "password", "shared", "apiKey") {
		if !fields[field] {
			t.Errorf("writer union omitted %s", field)
		}
	}
	if fields["email"] {
		t.Error("plain input field became encrypted")
	}
}

func TestLegacyPlaintextAuthSecretHasNarrowProviderAndFieldScope(t *testing.T) {
	token := "0123456789abcdef0123456789abcdef01234567"
	for _, field := range []string{AuthFieldAccessToken, AuthFieldRefreshToken, AuthFieldClientSecret} {
		if !IsLegacyPlaintextAuthSecret("strava", field, token) {
			t.Errorf("legacy Strava %s not recognized", field)
		}
	}
	for _, tc := range []struct{ plugin, field, value string }{
		{"hammerhead", AuthFieldAccessToken, token},
		{"strava", "password", token},
		{"strava", AuthFieldOAuthState, token},
		{"strava", AuthFieldClientSecret, token[:39]},
		{"strava", AuthFieldClientSecret, token + "0"},
		{"strava", AuthFieldClientSecret, "g" + token[1:]},
	} {
		if IsLegacyPlaintextAuthSecret(tc.plugin, tc.field, tc.value) {
			t.Errorf("unrelated credential admitted as legacy plaintext: %s/%s", tc.plugin, tc.field)
		}
	}
}
