package pluginsystem

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pocketbase/util"
)

func TestHostFailureBodyLimitsKeepStatusAndBackoff(t *testing.T) {
	useUnsafeTestHTTPClient(t)
	for _, status := range []int{200, 429, 503} {
		for _, incomplete := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/incomplete=%t", status, incomplete), func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if incomplete {
						conn, writer, err := w.(http.Hijacker).Hijack()
						if err != nil {
							t.Error(err)
							return
						}
						defer conn.Close()
						contentLength := 4096
						if status == 200 {
							contentLength = 32
						}
						_, _ = fmt.Fprintf(writer, "HTTP/1.1 %d Failure\r\nContent-Type: application/json\r\nRetry-After: 120\r\nContent-Length: %d\r\n\r\nincomplete-body-secret", status, contentLength)
						_ = writer.Flush()
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Retry-After", "120")
					w.WriteHeader(status)
					_, _ = w.Write([]byte(strings.Repeat("provider-body-secret", 100)))
				}))
				defer server.Close()
				spec := HostRequestSpec{Method: "GET", Target: RequestTarget{Type: "connector", Connector: "api", Path: "/v1"}, Expect: ResponseExpect{MaxBytes: 64}}
				data, _ := json.Marshal(spec)
				got := executeHostHTTPRequest(context.Background(), testHostManifest(t, server.URL), testHostPolicy(t, server.URL), data)
				if status == 200 {
					code := "invalid_output"
					if incomplete {
						code = "provider_unavailable"
					}
					if got.Error == nil || got.Error.Code != code {
						t.Fatalf("success body failure = %#v, want %s", got, code)
					}
					return
				}
				if got.Error != nil || got.Status != status {
					t.Fatalf("failure response lost status: %#v", got)
				}
				classified := ClassifyProviderError(OperationList, got.Status, got.HeaderValues["Retry-After"][0], "", time.Now())
				code := "provider_unavailable"
				if status == 429 {
					code = "rate_limited"
				}
				if classified.Code != code || classified.RetryAfterSeconds == nil || *classified.RetryAfterSeconds != 120 {
					t.Fatalf("classification lost: %#v", classified)
				}
				if len(got.BodyBase64) > 88 {
					t.Fatalf("error body exceeded 64 bytes: %d encoded bytes", len(got.BodyBase64))
				}
			})
		}
	}
}

func TestOAuthIncompleteFailureBodyKeepsStatus(t *testing.T) {
	for _, status := range []int{429, 503} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, writer, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			defer conn.Close()
			_, _ = fmt.Fprintf(writer, "HTTP/1.1 %d Failure\r\nRetry-After: 120\r\nContent-Length: 4096\r\n\r\nincomplete-secret", status)
			_ = writer.Flush()
		}))
		_, err := ExchangeOAuthToken(context.Background(), testHostManifest(t, server.URL), AuthContext{TokenURL: server.URL + "/v1"}, map[string]any{"clientId": "client-secret"}, map[string]string{"grant_type": "refresh_token", "refresh_token": "refresh-secret"})
		server.Close()
		if err == nil {
			t.Fatal("expected token rejection")
		}
		got := PublicPluginError(err)
		code := "provider_unavailable"
		if status == 429 {
			code = "rate_limited"
		}
		if got.Code != code || got.HTTPStatus != status || got.RetryAfterSeconds == nil || *got.RetryAfterSeconds != 120 {
			t.Fatalf("token HTTP metadata lost: %#v", got)
		}
	}
}

func TestHostPanicDropsPanicValue(t *testing.T) {
	original := newConnectorHTTPClient
	newConnectorHTTPClient = func(util.ConnectorHTTPPolicy, func(*http.Request, []*http.Request) error) (*http.Client, error) {
		panic("panic-secret-value")
	}
	t.Cleanup(func() { newConnectorHTTPClient = original })
	spec := HostRequestSpec{Method: "GET", Target: RequestTarget{Type: "connector", Connector: "api", Path: "/v1"}}
	data, _ := json.Marshal(spec)
	got := executeHostHTTPRequest(context.Background(), testHostManifest(t, "https://provider.test"), testHostPolicy(t, "https://provider.test"), data)
	if got.Error == nil || got.Error.Code != "runtime_error" {
		t.Fatalf("host panic = %#v", got)
	}
	encoded, _ := json.Marshal(got)
	if strings.Contains(string(encoded), "panic-secret-value") {
		t.Fatalf("panic leaked: %s", encoded)
	}
}

type errorContractPanicReader struct{}

func (errorContractPanicReader) Read([]byte) (int, error) { panic("panic-secret-value") }

func TestWorkerPanicDropsValueFromRPCAndStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunPluginWorker(context.Background(), errorContractPanicReader{}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("unexpected worker exit %d", code)
	}
	if strings.Contains(stderr.String(), "panic-secret-value") || !strings.Contains(stderr.String(), "runtime_error") {
		t.Fatalf("worker stderr: %s", stderr.String())
	}
	msg, err := readWorkerMessage(&stdout, defaultWorkerResponseMaxBytes)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := workerData[workerError](msg)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Message != "runtime_error" {
		t.Fatalf("worker RPC leaked: %#v", payload)
	}
}

func TestWorkerGuestLogCannotLeakResponseOrCredentials(t *testing.T) {
	var captured bytes.Buffer
	original := log.Writer()
	log.SetOutput(&captured)
	t.Cleanup(func() { log.SetOutput(original) })
	session := workerRuntimeSession{sessionID: "safe-session"}
	msg, err := workerMessageWithData(workerMessageHostLog, workerHostLog{Level: "warn", SessionID: "guest-secret-session", Message: "provider-body-secret https://user:secret@provider.test/?token=secret"})
	if err != nil {
		t.Fatal(err)
	}
	session.handleHostLog(msg)
	if strings.Contains(captured.String(), "secret") || !strings.Contains(captured.String(), "safe-session") {
		t.Fatalf("unsafe guest log: %s", captured.String())
	}
}
