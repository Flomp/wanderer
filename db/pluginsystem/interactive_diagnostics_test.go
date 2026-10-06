package pluginsystem

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"strings"
	"testing"
)

func TestHostAuthInjectionPreservesGrantClassification(t *testing.T) {
	for _, tc := range []struct{ name, auth, code, operation, cause string }{
		{"missing access token", "account", "invalid_grant", OperationOAuthRefresh, "host_auth_invalid"},
		{"unknown auth reference", "auth-context-secret", "invalid_request", OperationHostHTTP, "host_auth_invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manifest := testHostManifest(t, "https://provider-secret.test")
			manifest.Auth.Contexts = map[string]AuthContext{"account": {Type: AuthTypeOAuth2}}
			manifest.Permissions.Auth = []string{"account"}
			manifest.Permissions.Network.Connectors[0].Auth = []string{"account"}
			policy := testHostPolicy(t, "https://provider-secret.test")
			var output bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
			t.Cleanup(func() { slog.SetDefault(previous) })
			_, err := ExecuteHostRequest(context.Background(), manifest, policy, HostRequestSpec{Method: "GET", Auth: tc.auth, Target: RequestTarget{Type: "connector", Connector: "api", Path: "/v1", Query: []QueryParam{{Name: "token", Value: "query-secret"}}}}, HostRequestOptions{})
			if err == nil {
				t.Fatal("request reached the network without permitted credentials")
			}
			safe := PublicPluginError(err)
			if safe.Code != tc.code || safe.Operation != tc.operation {
				t.Fatalf("grant classification lost: %#v", safe)
			}
			var record map[string]any
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if record["code"] != tc.code || record["operation"] != tc.operation || record["cause"] != tc.cause {
				t.Fatalf("host diagnostic lost auth classification: %s", output.String())
			}
			if tc.code == "invalid_grant" && record["class"] != "authentication" {
				t.Fatalf("grant class: %s", output.String())
			}
			if strings.Contains(output.String(), "secret") {
				t.Fatalf("host diagnostic leaked source: %s", output.String())
			}
		})
	}
}

func TestGuestDiagnosticRetryAfterSurvivesWorkerRPC(t *testing.T) {
	for _, tc := range []struct {
		name  string
		retry int
		want  string
	}{
		{"zero", 0, "retry_after_seconds=0"},
		{"bounded", 90000, "retry_after_seconds=86400"},
		{"negative", -1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&output)
			t.Cleanup(func() { log.SetOutput(previous) })
			retry := tc.retry
			msg, err := workerMessageWithData(workerMessageHostLog, workerHostLog{Level: "warn", Message: "provider-body-secret", Event: "request_failed", SessionID: "guest-session-secret", Error: &PluginError{Code: "rate_limited", Message: "provider-token-secret", Operation: OperationDetail, HTTPStatus: 429, RetryAfterSeconds: &retry}})
			if err != nil {
				t.Fatal(err)
			}
			session := workerRuntimeSession{sessionID: "host-session"}
			session.handleHostLog(msg)
			got := output.String()
			for _, want := range []string{"event=request_failed", "code=rate_limited", "class=rate_limit", "operation=detail", "http_status=429", "host-session"} {
				if !strings.Contains(got, want) {
					t.Fatalf("lost %s: %s", want, got)
				}
			}
			if tc.want == "" {
				if strings.Contains(got, "retry_after_seconds") {
					t.Fatalf("negative delay retained: %s", got)
				}
			} else if !strings.Contains(got, tc.want) {
				t.Fatalf("lost retry metadata: %s", got)
			}
			if strings.Contains(got, "secret") {
				t.Fatalf("unsafe worker metadata: %s", got)
			}
		})
	}
}
