package pluginsystem

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Both guest runtimes call this same wire boundary. The lower-level executor
// must retain classification without logging before its caller has context.
func TestHostHTTPDiagnosticLogsOnceAtGuestBoundary(t *testing.T) {
	cases := []struct {
		name               string
		payload            []byte
		code, cause, class string
		status             int
	}{
		{"malformed payload", []byte(`{"target":"payload-secret`), "invalid_request", "input_encoding", "configuration", 0},
		{"policy rejection", []byte(`{"method":"GET","target":{"type":"connector","connector":"connector-secret","path":"/v1"}}`), "invalid_request", "request_policy_rejected", "configuration", 0},
		{"response media rejection", nil, "invalid_output", "response_policy_rejected", "runtime", 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			useUnsafeTestHTTPClient(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte("provider-body-secret"))
			}))
			t.Cleanup(server.Close)
			manifest := testHostManifest(t, server.URL)
			policy := testHostPolicy(t, server.URL)
			data := tc.payload
			if data == nil {
				spec := HostRequestSpec{Method: "GET", Target: RequestTarget{Type: "connector", Connector: "api", Path: "/v1", Query: []QueryParam{{Name: "token", Value: "query-secret"}}}, Headers: map[string]string{"Authorization": "Bearer header-secret"}}
				var err error
				data, err = json.Marshal(spec)
				if err != nil {
					t.Fatal(err)
				}
			}
			var output bytes.Buffer
			original := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
			t.Cleanup(func() { slog.SetDefault(original) })
			var spec HostRequestSpec
			if json.Unmarshal(data, &spec) == nil {
				_, err := ExecuteHostRequest(context.Background(), manifest, policy, spec, HostRequestOptions{})
				if err == nil {
					t.Fatal("expected classified host failure")
				}
				safe := PublicPluginError(err)
				if safe.Code != tc.code || safe.HTTPStatus != tc.status {
					t.Fatalf("executor changed classification: %#v", safe)
				}
				if output.Len() != 0 {
					t.Fatalf("executor logged prematurely: %s", output.String())
				}
			}
			response := executeHostHTTPRequest(context.Background(), manifest, policy, data)
			if response.Error == nil || response.Error.Code != tc.code || response.Error.Operation != OperationHostHTTP || response.Error.HTTPStatus != tc.status {
				t.Fatalf("wire classification changed: %#v", response.Error)
			}
			lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
			if len(lines) != 1 {
				t.Fatalf("guest boundary logged %d times: %s", len(lines), output.String())
			}
			var record map[string]any
			if err := json.Unmarshal(lines[0], &record); err != nil {
				t.Fatal(err)
			}
			for key, want := range map[string]any{"plugin": manifest.ID, "code": tc.code, "cause": tc.cause, "class": tc.class, "operation": OperationHostHTTP, "http_status": float64(tc.status)} {
				if record[key] != want {
					t.Errorf("diagnostic %s=%#v, want %#v: %s", key, record[key], want, output.String())
				}
			}
			for _, secret := range []string{"payload-secret", "connector-secret", "provider-body-secret", "query-secret", "header-secret", server.URL} {
				if strings.Contains(output.String(), secret) {
					t.Errorf("diagnostic leaked %q: %s", secret, output.String())
				}
			}
		})
	}
}

func TestHostHTTPExecutorNormalizesNetworkFailureWithoutLogging(t *testing.T) {
	useUnsafeTestHTTPClient(t)
	provider := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	provider.Close()
	var output bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })
	manifest := testHostManifest(t, provider.URL)
	policy := testHostPolicy(t, provider.URL)
	spec := HostRequestSpec{Method: "GET", Target: RequestTarget{Type: "connector", Connector: "api", Path: "/v1", Query: []QueryParam{{Name: "token", Value: "query-secret"}}}}
	_, err := ExecuteHostRequest(context.Background(), manifest, policy, spec, HostRequestOptions{})
	if err == nil || PublicPluginError(err).Code != "provider_unavailable" {
		t.Fatalf("lost network classification: %v", err)
	}
	if output.Len() != 0 {
		t.Fatalf("executor logged before caller: %s", output.String())
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	response := executeHostHTTPRequest(context.Background(), manifest, policy, data)
	if response.Error == nil || response.Error.Code != "provider_unavailable" {
		t.Fatalf("lost guest network failure: %#v", response.Error)
	}
	lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
	if len(lines) != 1 {
		t.Fatalf("guest boundary logged %d times: %s", len(lines), output.String())
	}
	if strings.Contains(output.String(), "query-secret") || strings.Contains(output.String(), provider.URL) {
		t.Fatalf("network diagnostic leaked source: %s", output.String())
	}
}
