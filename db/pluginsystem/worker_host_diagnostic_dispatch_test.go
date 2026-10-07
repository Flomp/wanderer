package pluginsystem

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sync/semaphore"
)

const workerDiagnosticDispatchHelperEnv = "WANDERER_TEST_DIAGNOSTIC_DISPATCH"

// Fault injection happens only on the child's output pipe, after it receives a
// real call_export. The parent uses its production Call/call dispatch, framing,
// fatal-session guard, process teardown and acquired-slot lifecycle unchanged.
func TestWorkerDiagnosticDispatchProcessHelper(t *testing.T) {
	mode := os.Getenv(workerDiagnosticDispatchHelperEnv)
	if mode == "" {
		return
	}
	if err := runWorkerDiagnosticDispatchChild(mode); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func runWorkerDiagnosticDispatchChild(mode string) error {
	for {
		call, err := readWorkerMessage(os.Stdin, defaultWorkerRequestMaxBytes)
		if err != nil {
			return err
		}
		if call.Type == workerMessageShutdown {
			return nil
		}
		if call.Type != workerMessageCallExport {
			return fmt.Errorf("unexpected input frame")
		}
		request, err := workerData[workerCallExport](call)
		if err != nil {
			return err
		}
		valid, err := workerMessageWithData(workerMessageHostDiagnostic, workerHostDiagnostic{Kind: workerHostDiagnosticHTTPInputInvalid})
		if err != nil {
			return err
		}
		if mode == "valid_repeated" {
			for range 2 {
				if err := writeWorkerMessage(os.Stdout, defaultWorkerResponseMaxBytes, valid); err != nil {
					return err
				}
			}
			result, err := workerMessageWithData(workerMessageCallResult, workerCallResult{OutputBase64: request.InputBase64})
			if err != nil {
				return err
			}
			if err := writeWorkerMessage(os.Stdout, defaultWorkerResponseMaxBytes, result); err != nil {
				return err
			}
			continue
		}
		if mode == "valid_then_unknown" {
			if err := writeWorkerMessage(os.Stdout, defaultWorkerResponseMaxBytes, valid); err != nil {
				return err
			}
		}
		if mode == "malformed_frame" {
			// A broken outer frame exercises readWorkerMessage's decoding path.
			payload := []byte(`{"type":"host_diagnostic","data":{"kind":"private-payload-sentinel`)
			var header [4]byte
			binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
			if err := writeAll(os.Stdout, header[:]); err != nil {
				return err
			}
			if err := writeAll(os.Stdout, payload); err != nil {
				return err
			}
		} else {
			var payload json.RawMessage
			switch mode {
			case "malformed_payload":
				// The outer frame is valid, but Data cannot decode into the
				// private diagnostic structure. This reaches its real dispatch.
				payload = json.RawMessage(`"private-payload-sentinel"`)
			case "missing_kind":
				payload = json.RawMessage(`{"message":"private-payload-sentinel"}`)
			case "unknown_kind", "valid_then_unknown":
				payload = json.RawMessage(`{"kind":"private-payload-sentinel"}`)
			default:
				return fmt.Errorf("unknown helper mode")
			}
			if err := writeWorkerMessage(os.Stdout, defaultWorkerResponseMaxBytes, workerMessage{Type: workerMessageHostDiagnostic, Data: payload}); err != nil {
				return err
			}
		}
		// Neither closing stdin nor waiting for Close should be needed to stop
		// an invalid worker. Only the production fatal dispatch's kill ends us.
		time.Sleep(time.Minute)
		return nil
	}
}

func TestWorkerInvalidHostDiagnosticsAbortRealDispatch(t *testing.T) {
	for _, mode := range []string{"malformed_payload", "missing_kind", "unknown_kind", "valid_then_unknown", "malformed_frame"} {
		t.Run(mode, func(t *testing.T) {
			session := newWorkerDiagnosticDispatchProcess(t, mode)
			var output bytes.Buffer
			setPluginTestDefaultLogger(t, slog.New(slog.NewJSONHandler(&output, nil)))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			result, err := session.Call(ctx, "diagnostic_dispatch_v1", []byte("call-input-sentinel"))
			if len(result) != 0 || !IsRuntimeSessionFatalError(err) {
				t.Fatalf("invalid private frame was not fatal: output=%q error=%v", result, err)
			}
			if err.Error() != "runtime_error" {
				t.Fatalf("invalid private frame exposed its cause: %v", err)
			}
			public := PublicPluginError(err)
			if public.Code != "runtime_error" || public.Message != "runtime_error" || public.Operation != OperationRuntime {
				t.Errorf("invalid private frame lost its safe runtime classification: %#v", public)
			}
			session.mu.Lock()
			fatal, reason := session.fatal, session.fatalMsg
			session.mu.Unlock()
			wantReason := "invalid host diagnostic"
			if mode == "malformed_frame" {
				wantReason = "read worker message failed"
			}
			if !fatal || reason != wantReason {
				t.Errorf("real dispatch did not mark its session fatal: fatal=%v reason=%q", fatal, reason)
			}
			// Wait before Close: closing the session cannot hide a missing kill.
			waited := make(chan error, 1)
			go func() { waited <- session.wait() }()
			select {
			case waitErr := <-waited:
				status, ok := session.cmd.ProcessState.Sys().(syscall.WaitStatus)
				if waitErr == nil || !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
					t.Errorf("fatal dispatch did not kill/reap child: error=%v state=%v", waitErr, session.cmd.ProcessState)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("fatal dispatch did not kill its child before Close")
			}
			if _, err := session.Call(ctx, "diagnostic_dispatch_v1", nil); !IsRuntimeSessionFatalError(err) || err.Error() != "runtime_error" {
				t.Fatalf("fatal worker was reusable or leaked its cause: %v", err)
			}
			if mode == "valid_then_unknown" {
				record := singlePluginDiagnosticRecord(t, &output)
				if record["msg"] != "plugin host request failed" || record["cause"] != "input_encoding" {
					t.Errorf("valid preceding frame lost its trusted diagnostic: %s", output.String())
				}
			} else if output.Len() != 0 {
				t.Errorf("invalid frame acquired a trusted diagnostic: %s", output.String())
			}
			encoded, marshalErr := json.Marshal(public)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			for _, text := range []string{string(encoded), output.String(), session.stderr.String()} {
				if strings.Contains(text, "sentinel") {
					t.Errorf("private frame escaped into public failure/log/stderr: %q", text)
				}
			}
			if err := session.Close(ctx); err != nil {
				t.Fatal(err)
			}
			if !session.slot.TryAcquire(1) {
				t.Fatal("closing fatal worker did not release its acquired slot")
			}
			session.slot.Release(1)
		})
	}
}

func TestWorkerRepeatedHostDiagnosticsContinueRealDispatch(t *testing.T) {
	session := newWorkerDiagnosticDispatchProcess(t, "valid_repeated")
	var output bytes.Buffer
	setPluginTestDefaultLogger(t, slog.New(slog.NewJSONHandler(&output, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, input := range []string{"first", "second"} {
		result, err := session.Call(ctx, "diagnostic_dispatch_v1", []byte(input))
		if err != nil || string(result) != input {
			t.Fatalf("valid diagnostics interrupted call_result/session reuse: output=%q error=%v", result, err)
		}
	}
	lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
	if len(lines) != 4 {
		t.Fatalf("want two trusted diagnostics per call, got %d: %s", len(lines), output.String())
	}
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatal(err)
		}
		for key, want := range map[string]any{"msg": "plugin host request failed", "code": "invalid_request", "class": "configuration", "operation": OperationHostHTTP, "cause": "input_encoding", "plugin": "dispatch-test", "session": "dispatch-session"} {
			if record[key] != want {
				t.Errorf("trusted repeated diagnostic %s=%#v, want %#v: %s", key, record[key], want, line)
			}
		}
	}
}

func newWorkerDiagnosticDispatchProcess(t *testing.T, mode string) *workerRuntimeSession {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	processCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	cmd := exec.CommandContext(processCtx, executable, "-test.run=^TestWorkerDiagnosticDispatchProcessHelper$")
	cmd.Env = append(childEnvWithout("EXTISM_ENABLE_WASI_OUTPUT"), workerDiagnosticDispatchHelperEnv+"="+mode)
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
	slot := semaphore.NewWeighted(1)
	if err := slot.Acquire(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	session := &workerRuntimeSession{
		plugin: LocalPlugin{Manifest: Manifest{ID: "dispatch-test"}}, sessionID: "dispatch-session",
		cmd: cmd, stdin: stdin, stdout: stdout, stderr: stderr, slot: slot,
		requestMaxBytes: defaultWorkerRequestMaxBytes, responseMaxBytes: defaultWorkerResponseMaxBytes,
		exportTimeout: 5 * time.Second,
	}
	t.Cleanup(func() {
		defer cancel()
		defer stdout.Close()
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer closeCancel()
		if err := session.Close(closeCtx); err != nil {
			t.Errorf("dispatch worker cleanup failed: %v", err)
		}
	})
	return session
}
