package pluginsystem

import (
	"testing"
	"time"
)

func TestLocalConnectionStatusUsesOAuthGrantProvenance(t *testing.T) {
	now := time.Now()
	context := AuthContext{Type: AuthTypeOAuth2, Fields: []string{"clientId", "clientSecret"}, AuthorizationURL: "https://example.test/authorize", TokenURL: "https://example.test/token", Refresh: &AuthRefresh{Mode: AuthRefreshModeHost}}
	for _, test := range []struct {
		name       string
		multiple   bool
		provenance any
		set        bool
		want       string
	}{
		{name: "matching grant", multiple: true, provenance: "selected", set: true, want: "configured"},
		{name: "foreign grant", multiple: true, provenance: "other", set: true, want: "needs_auth"},
		{name: "foreign grant in sole context", provenance: "old-context", set: true, want: "needs_auth"},
		{name: "unattributed multi-context grant", multiple: true, want: "needs_auth"},
		{name: "legacy unattributed sole grant", want: "configured"},
		{name: "legacy empty provenance sole grant", provenance: "", set: true, want: "configured"},
		{name: "nonstring provenance", provenance: 42, set: true, want: "needs_auth"},
		{name: "null provenance", provenance: nil, set: true, want: "needs_auth"},
	} {
		t.Run(test.name, func(t *testing.T) {
			manifest := Manifest{Auth: AuthManifest{Contexts: map[string]AuthContext{"selected": context}}}
			if test.multiple {
				manifest.Auth.Contexts["other"] = context
			}
			auth := map[string]any{"accessToken": "unread-but-present-token", "refreshToken": "unread-but-present-refresh", "expiresAt": now.Add(time.Hour).Unix(), "clientId": "client", "clientSecret": "unread-secret"}
			if test.set {
				auth[AuthFieldOAuthContext] = test.provenance
			}
			if got := LocalConnectionStatus(manifest, "selected", auth, true, now); got != test.want {
				t.Fatalf("local status = %s, want %s", got, test.want)
			}
			if got := LocalConnectionStatus(manifest, "selected", auth, false, now); got != "disabled" {
				t.Fatalf("disabled grant status = %s", got)
			}
			auth["expiresAt"] = now.Add(-time.Hour).Unix()
			if got := LocalConnectionStatus(manifest, "selected", auth, true, now); got != test.want {
				t.Fatalf("refresh-ready status = %s, want %s", got, test.want)
			}
		})
	}
}
