package pluginsystem

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	wasmbinary "github.com/tetratelabs/wabin/binary"
	"github.com/tetratelabs/wabin/wasm"
	"golang.org/x/sync/semaphore"
)

// This is the actual worker entrypoint in a separate OS process. Keeping the
// helper behind both a selected test and a private environment flag prevents
// ordinary package tests from becoming workers or writing protocol stdout.
func TestGuestLogWorkerProcessHelper(t *testing.T) {
	if os.Getenv("WANDERER_TEST_GUEST_LOG_WORKER") != "1" {
		return
	}
	os.Exit(RunPluginWorker(context.Background(), os.Stdin, os.Stdout, os.Stderr))
}

// This covers the production WASM host callback in worker_process.go, child
// framing and the parent's Call/handleHostLog path. A hand-built host_log frame
// cannot detect a child callback that forgets to forward Event or Error.
func TestGuestLogEventSurvivesActualWorkerProcess(t *testing.T) {
	wasmPath := filepath.Join(t.TempDir(), "log.wasm")
	if err := os.WriteFile(wasmPath, guestLogWASMFixture(), 0600); err != nil {
		t.Fatal(err)
	}
	session := newGuestLogWorkerProcess(t, wasmPath)
	var captured bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&captured)
	t.Cleanup(func() { log.SetOutput(previous) })

	const privateText = "provider-body-sentinel https://user:credential-sentinel@provider-sentinel.test/?token=query-sentinel"
	cases := []struct {
		name  string
		entry HostLogEntry
		event string
		typed bool
		retry string
	}{
		{name: "zero", entry: guestLogRetryEntry(privateText, bundledRetrySeconds(0)), event: "request_failed", typed: true, retry: "retry_after_seconds=0"},
		{name: "bounded", entry: guestLogRetryEntry(privateText, bundledRetrySeconds(90000)), event: "request_failed", typed: true, retry: "retry_after_seconds=86400"},
		{name: "negative", entry: guestLogRetryEntry(privateText, bundledRetrySeconds(-1)), event: "request_failed", typed: true},
		{name: "nil retry", entry: guestLogRetryEntry(privateText, nil), event: "request_failed", typed: true},
		{name: "event without error", entry: HostLogEntry{Level: "info", Message: privateText, Event: "request_completed"}, event: "request_completed"},
		{name: "legacy free text", entry: HostLogEntry{Level: "warn", Message: privateText}, event: "guest_message_redacted"},
		{name: "untrusted metadata", entry: HostLogEntry{Level: "warn", Message: privateText, Event: "event-sentinel", Error: &PluginError{Code: "code-sentinel", Message: privateText, Operation: "operation-sentinel", HTTPStatus: 999, RetryAfterSeconds: bundledRetrySeconds(-1)}}, event: "guest_message_redacted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			captured.Reset()
			payload, err := json.Marshal(tc.entry)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := session.Call(ctx, "log_event_v1", payload); err != nil {
				t.Fatalf("WASM/child/parent log call failed: %v; child stderr %q", err, session.stderr.String())
			}
			got := captured.String()
			if strings.Count(got, "plugin log [") != 1 {
				t.Fatalf("expected one parent diagnostic: %q", got)
			}
			for _, want := range []string{"session host-log-test:", "event=" + tc.event} {
				if !strings.Contains(got, want) {
					t.Fatalf("child forwarding lost %s: %q", want, got)
				}
			}
			if tc.typed {
				for _, want := range []string{"code=rate_limited", "class=rate_limit", "operation=detail", "http_status=429"} {
					if !strings.Contains(got, want) {
						t.Fatalf("child forwarding lost %s: %q", want, got)
					}
				}
			} else if tc.entry.Error == nil {
				if strings.Contains(got, "code=") || strings.Contains(got, "http_status=") {
					t.Fatalf("a log without an error acquired failure metadata: %q", got)
				}
			} else {
				for _, want := range []string{"code=plugin_error", "class=runtime", "http_status=0"} {
					if !strings.Contains(got, want) {
						t.Fatalf("untrusted child metadata was not normalized: %q", got)
					}
				}
			}
			if tc.retry == "" {
				if strings.Contains(got, "retry_after_seconds") {
					t.Fatalf("missing/negative retry was retained: %q", got)
				}
			} else if !strings.Contains(got, tc.retry) {
				t.Fatalf("child forwarding lost retry metadata: %q", got)
			}
			for _, text := range []string{got, session.stderr.String()} {
				if strings.Contains(text, "sentinel") || strings.Contains(text, "https://") || strings.Contains(text, privateText) {
					t.Fatalf("guest data escaped into logs/stderr: %q", text)
				}
			}
		})
	}
}

func guestLogRetryEntry(message string, retry *int) HostLogEntry {
	return HostLogEntry{Level: "warn", Message: message, Event: "request_failed", Error: &PluginError{
		Code: "rate_limited", Message: message, Operation: OperationDetail, HTTPStatus: 429, RetryAfterSeconds: retry,
	}}
}

func newGuestLogWorkerProcess(t *testing.T, wasmPath string) *workerRuntimeSession {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	processCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	cmd := exec.CommandContext(processCtx, executable, "-test.run=^TestGuestLogWorkerProcessHelper$")
	cmd.Env = append(childEnvWithout("EXTISM_ENABLE_WASI_OUTPUT"), "WANDERER_TEST_GUEST_LOG_WORKER=1")
	stderr := &boundedWorkerBuffer{limit: defaultWorkerCapturedStderrBytes}
	cmd.Stderr = stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		cancel()
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		cancel()
		t.Fatal(err)
	}
	// Close uses the real session lifecycle, including its acquired slot.
	slot := semaphore.NewWeighted(1)
	if err := slot.Acquire(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	session := &workerRuntimeSession{
		plugin:    LocalPlugin{Manifest: Manifest{ID: "worker-log-test"}, WASMPath: wasmPath},
		sessionID: "host-log-test", cmd: cmd, stdin: stdin, stdout: stdout, stderr: stderr,
		requestMaxBytes: defaultWorkerRequestMaxBytes, responseMaxBytes: defaultWorkerResponseMaxBytes,
		exportTimeout: 5 * time.Second, slot: slot,
	}
	t.Cleanup(func() {
		defer cancel()
		defer stdout.Close()
		closeCtx, cancelClose := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelClose()
		if err := session.Close(closeCtx); err != nil {
			t.Errorf("worker did not exit cleanly: %v; stderr %q", err, stderr.String())
		}
	})
	return session
}

// This tiny guest calls wanderer.log with its Extism input allocation and
// returns success. It uses the same import/ABI as SDK LogEvent, with arbitrary
// wire payloads to cover normalization of untrusted legacy plugins. No TinyGo,
// downloaded fixtures or optional bundle environment variable are required.
//
// (module
//
//	(import "extism:host/env" "input_offset" (func (result i64)))
//	(import "wanderer" "log" (func (param i64)))
//	(func (export "log_event_v1") (result i32)
//	  call 0
//	  call 1
//	  i32.const 0))
func guestLogWASMFixture() []byte {
	return wasmbinary.EncodeModule(&wasm.Module{
		TypeSection: []*wasm.FunctionType{
			{Results: []wasm.ValueType{wasm.ValueTypeI64}},
			{Params: []wasm.ValueType{wasm.ValueTypeI64}},
			{Results: []wasm.ValueType{wasm.ValueTypeI32}},
		},
		ImportSection: []*wasm.Import{
			{Type: wasm.ExternTypeFunc, Module: "extism:host/env", Name: "input_offset", DescFunc: 0},
			{Type: wasm.ExternTypeFunc, Module: "wanderer", Name: "log", DescFunc: 1},
		},
		FunctionSection: []wasm.Index{2},
		ExportSection:   []*wasm.Export{{Name: "log_event_v1", Type: wasm.ExternTypeFunc, Index: 2}},
		CodeSection:     []*wasm.Code{{Body: []byte{wasm.OpcodeCall, 0, wasm.OpcodeCall, 1, wasm.OpcodeI32Const, 0, wasm.OpcodeEnd}}},
	})
}
