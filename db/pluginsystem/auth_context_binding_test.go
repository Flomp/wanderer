package pluginsystem

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestWorkerHostHTTPUsesTrustedActiveContextForSharedCredentialNames(t *testing.T) {
	useUnsafeTestHTTPClient(t)
	for _, tc := range []struct {
		name, selected, requested string
		bound, allowed            bool
	}{
		{"active_context_allowed", "active", "active", true, true},
		{"inactive_alias_rejected", "active", "inactive", true, false},
		{"unresolved_selection_rejected", "", "active", true, false},
		{"unbound_auth_rejected", "", "inactive", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32
			var observed atomic.Value
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				observed.Store(r.Header.Get("Authorization"))
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer server.Close()
			manifest := testHostManifest(t, server.URL)
			manifest.Auth.Contexts = map[string]AuthContext{"active": {Type: AuthTypeBearer, SecretField: "token"}, "inactive": {Type: AuthTypeBearer, SecretField: "token"}}
			manifest.Permissions.Auth = []string{"active", "inactive"}
			manifest.Permissions.Network.Connectors[0].Auth = []string{"active", "inactive"}
			activeAuth, err := ActiveAuth(manifest, map[string]any{"token": "active-context-secret"}, "active")
			if err != nil {
				t.Fatal(err)
			}
			policy := testHostPolicy(t, server.URL).WithHostAuth(activeAuth)
			if tc.bound {
				policy = policy.WithHostAuth(activeAuth, tc.selected)
			}
			payload, err := json.Marshal(HostRequestSpec{Method: http.MethodGet, Auth: tc.requested, Target: RequestTarget{Type: "connector", Connector: "api", Path: "/v1"}, Expect: ResponseExpect{ContentTypes: []string{"application/json"}, MaxBytes: 1024}})
			if err != nil {
				t.Fatal(err)
			}
			response := executeHostHTTPRequest(context.Background(), manifest, policy, payload)
			if tc.allowed {
				if response.Error != nil || requests.Load() != 1 || observed.Load() != "Bearer active-context-secret" {
					t.Fatalf("active request rejected: %#v", response)
				}
			} else if response.Error == nil || requests.Load() != 0 {
				t.Fatalf("inactive/unbound request reached provider: %#v", response)
			}
		})
	}
}
