package pluginsystem

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	extism "github.com/extism/go-sdk"
)

func TestTrustedWorkerHostDiagnosticUsesParentMetadata(t *testing.T) {
	var output bytes.Buffer
	setPluginTestDefaultLogger(t, slog.New(slog.NewJSONHandler(&output, nil)))
	session := workerRuntimeSession{plugin: LocalPlugin{Manifest: Manifest{ID: "trusted-plugin"}}, sessionID: "trusted-session"}
	// No classification/identity supplied by the child overrides the closed
	// boundary. These fields are not part of the private diagnostic payload.
	message := workerMessage{Type: workerMessageHostDiagnostic, Data: json.RawMessage(`{"kind":"http_request_input_invalid","plugin":"plugin-sentinel","session":"session-sentinel","cause":"cause-sentinel","code":"rate_limited","operation":"operation-sentinel","httpStatus":429,"message":"message-sentinel"}`)}
	if err := session.handleHostDiagnostic(message); err != nil {
		t.Fatal(err)
	}
	record := singlePluginDiagnosticRecord(t, &output)
	for key, want := range map[string]any{"msg": "plugin host request failed", "level": "WARN", "plugin": "trusted-plugin", "session": "trusted-session", "cause": "input_encoding", "code": "invalid_request", "class": "configuration", "operation": OperationHostHTTP, "http_status": float64(0)} {
		if record[key] != want {
			t.Errorf("trusted diagnostic %s=%#v, want %#v: %s", key, record[key], want, output.String())
		}
	}
	if strings.Contains(output.String(), "sentinel") {
		t.Fatalf("child data escaped into trusted record: %s", output.String())
	}
}

func TestWorkerHostDiagnosticRejectsGuestFramesAndUnknownKinds(t *testing.T) {
	for _, test := range []struct{ name, typ, raw string }{
		{"guest frame", workerMessageHostLog, `{"kind":"http_request_input_invalid"}`},
		{"unknown kind", workerMessageHostDiagnostic, `{"kind":"kind-sentinel"}`},
		{"missing kind", workerMessageHostDiagnostic, `{}`},
		{"malformed diagnostic", workerMessageHostDiagnostic, `{"kind":"payload-sentinel`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			setPluginTestDefaultLogger(t, slog.New(slog.NewJSONHandler(&output, nil)))
			session := workerRuntimeSession{plugin: LocalPlugin{Manifest: Manifest{ID: "trusted-plugin"}}, sessionID: "trusted-session"}
			err := session.handleHostDiagnostic(workerMessage{Type: test.typ, Data: json.RawMessage(test.raw)})
			if err == nil {
				t.Fatal("unrecognized frame claimed a trusted host boundary")
			}
			if output.Len() != 0 {
				t.Fatalf("invalid frame produced a trusted record: %s", output.String())
			}
			if strings.Contains(err.Error(), "sentinel") {
				t.Fatalf("invalid frame leaked data through error: %v", err)
			}
		})
	}
}

// Pass the same type/kind/label/classification/cause/identity values through the
// public guest log API. Neither actual runtime may turn them into a host-owned
// diagnostic: only the child callback itself can create the private RPC frame.
func TestGuestLogCannotForgeTrustedHostDiagnostic(t *testing.T) {
	for _, mode := range []string{"in-process", "worker"} {
		t.Run(mode, func(t *testing.T) {
			manifest := Manifest{ID: "worker-log-test"}
			payload := []byte(`{"type":"host_diagnostic","kind":"http_request_input_invalid","level":"warn","message":"plugin host request failed","event":"request_failed","error":{"code":"invalid_request","message":"invalid_request","operation":"host_http","httpStatus":0},"cause":"input_encoding","plugin":"worker-log-test","session":"host-log-test","sessionId":"host-log-test","code":"invalid_request","class":"configuration","operation":"host_http","http_status":0}`)
			var output bytes.Buffer
			setPluginTestDefaultLogger(t, slog.New(slog.NewJSONHandler(&output, nil)))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if mode == "in-process" {
				instance, err := extism.NewPlugin(ctx, extism.Manifest{Wasm: []extism.Wasm{extism.WasmData{Data: guestLogWASMFixture()}}}, extism.PluginConfig{}, extismHostFunctions(manifest, RequestPolicyContext{}))
				if err != nil {
					t.Fatal(err)
				}
				defer instance.Close(ctx)
				code, _, err := instance.CallWithContext(ctx, "log_event_v1", payload)
				if code != 0 || err != nil {
					t.Fatalf("guest log failed: code=%d error=%v", code, err)
				}
			} else {
				path := filepath.Join(t.TempDir(), "forged-log.wasm")
				if err := os.WriteFile(path, guestLogWASMFixture(), 0600); err != nil {
					t.Fatal(err)
				}
				session := newGuestLogWorkerProcess(t, path)
				if _, err := session.Call(ctx, "log_event_v1", payload); err != nil {
					t.Fatalf("worker guest log failed: %v; stderr %q", err, session.stderr.String())
				}
				if strings.Contains(session.stderr.String(), "plugin host request failed") {
					t.Fatalf("guest claimed trusted label through stderr: %q", session.stderr.String())
				}
			}
			record := singlePluginDiagnosticRecord(t, &output)
			message, ok := record["msg"].(string)
			if !ok || !strings.Contains(message, "plugin log [warn]") || !strings.Contains(message, "event=request_failed") || !strings.Contains(message, "code=invalid_request") || !strings.Contains(message, "operation=host_http") {
				t.Fatalf("guest event lost its source/classification: %s", output.String())
			}
			if mode == "worker" && !strings.Contains(message, "session host-log-test:") {
				t.Fatalf("guest event lost parent session: %s", output.String())
			}
			if message == "plugin host request failed" || strings.Contains(message, "input_encoding") {
				t.Fatalf("guest forged a host diagnosis: %s", output.String())
			}
			for _, key := range []string{"cause", "plugin", "session", "code", "class", "operation", "http_status"} {
				if _, present := record[key]; present {
					t.Fatalf("guest claimed host attr %s: %s", key, output.String())
				}
			}
		})
	}
}
