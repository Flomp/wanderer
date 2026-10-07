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
	wasmbinary "github.com/tetratelabs/wabin/binary"
	"github.com/tetratelabs/wabin/wasm"
)

// Exercise the actual Extism callback, not a fabricated host-response packet.
// The guest passes an address outside linear memory and returns the host's JSON.
func TestHostHTTPRequestInvalidPointerLogsClassifiedFailure(t *testing.T) {
	manifest := Manifest{ID: "pointer-http-test"}
	var output bytes.Buffer
	setPluginTestDefaultLogger(t, slog.New(slog.NewJSONHandler(&output, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	plugin, err := extism.NewPlugin(ctx, extism.Manifest{Wasm: []extism.Wasm{extism.WasmData{Data: invalidHTTPPointerWASMFixture()}}}, extism.PluginConfig{}, extismHostFunctions(manifest, RequestPolicyContext{}))
	if err != nil {
		t.Fatal(err)
	}
	defer plugin.Close(ctx)
	code, data, err := plugin.CallWithContext(ctx, "http_invalid_pointer_v1", []byte("guest-payload-sentinel"))
	if err != nil || code != 0 {
		t.Fatalf("callback did not return a safe response: code=%d error=%v", code, err)
	}
	assertInvalidHTTPPointerResponse(t, data)
	record := singlePluginDiagnosticRecord(t, &output)
	for key, want := range map[string]any{"msg": "plugin host request failed", "level": "WARN", "plugin": manifest.ID, "code": "invalid_request", "class": "configuration", "cause": "input_encoding", "operation": OperationHostHTTP, "http_status": float64(0)} {
		if record[key] != want {
			t.Errorf("pointer diagnostic %s=%#v, want %#v: %s", key, record[key], want, output.String())
		}
	}
	for _, secret := range []string{"sentinel", "invalid memory block", "18446744073709551615", "4294967295"} {
		if strings.Contains(output.String(), secret) {
			t.Errorf("pointer diagnostic leaked %q: %s", secret, output.String())
		}
	}
}

func TestHostHTTPRequestInvalidPointerSurvivesActualWorkerProcess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "http-pointer.wasm")
	if err := os.WriteFile(path, invalidHTTPPointerWASMFixture(), 0600); err != nil {
		t.Fatal(err)
	}
	session := newGuestLogWorkerProcess(t, path)
	var output bytes.Buffer
	setPluginTestDefaultLogger(t, slog.New(slog.NewJSONHandler(&output, nil)))
	// The invalid request is recoverable: each call returns its wire failure and
	// the same worker remains reusable, logging exactly once for each rejection.
	for i := 0; i < 2; i++ {
		output.Reset()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		data, err := session.Call(ctx, "http_invalid_pointer_v1", []byte("guest-payload-sentinel"))
		cancel()
		if err != nil {
			t.Fatalf("worker pointer callback failed: %v; stderr %q", err, session.stderr.String())
		}
		assertInvalidHTTPPointerResponse(t, data)
		got := output.String()
		record := singlePluginDiagnosticRecord(t, &output)
		for key, want := range map[string]any{"msg": "plugin host request failed", "level": "WARN", "plugin": session.plugin.Manifest.ID, "session": session.sessionID,
			"code": "invalid_request", "class": "configuration", "cause": "input_encoding", "operation": OperationHostHTTP, "http_status": float64(0)} {
			if record[key] != want {
				t.Errorf("worker host diagnostic %s=%#v, want %#v: %s", key, record[key], want, got)
			}
		}
		if strings.Contains(got, "plugin log [") {
			t.Errorf("trusted host failure was labeled as a guest event: %s", got)
		}
		for _, text := range []string{got, session.stderr.String()} {
			if strings.Contains(text, "sentinel") || strings.Contains(text, "invalid memory block") {
				t.Fatalf("guest memory error escaped: %q", text)
			}
		}
	}
}

func assertInvalidHTTPPointerResponse(t *testing.T, data []byte) {
	t.Helper()
	var response hostHTTPResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("invalid host response: %s: %v", data, err)
	}
	if response.Error == nil || response.Error.Code != "invalid_request" || response.Error.Message != "invalid_request" || response.Error.Operation != OperationHostHTTP || response.Error.HTTPStatus != 0 {
		t.Fatalf("invalid pointer lost safe classification: %s", data)
	}
	if strings.Contains(string(data), "sentinel") || strings.Contains(string(data), "invalid memory block") {
		t.Fatalf("host response leaked guest memory details: %s", data)
	}
}

// (module
//
//	(import "wanderer" "http_request" (func (param i64) (result i64)))
//	(import "extism:host/env" "length" (func (param i64) (result i64)))
//	(import "extism:host/env" "output_set" (func (param i64 i64)))
//	(func (export "http_invalid_pointer_v1") (result i32) (local i64)
//	  i64.const -1 call 0 local.tee 0 local.get 0 call 1 call 2 i32.const 0))
func invalidHTTPPointerWASMFixture() []byte {
	return wasmbinary.EncodeModule(&wasm.Module{
		TypeSection:     []*wasm.FunctionType{{Params: []wasm.ValueType{wasm.ValueTypeI64}, Results: []wasm.ValueType{wasm.ValueTypeI64}}, {Params: []wasm.ValueType{wasm.ValueTypeI64, wasm.ValueTypeI64}}, {Results: []wasm.ValueType{wasm.ValueTypeI32}}},
		ImportSection:   []*wasm.Import{{Type: wasm.ExternTypeFunc, Module: "wanderer", Name: "http_request", DescFunc: 0}, {Type: wasm.ExternTypeFunc, Module: "extism:host/env", Name: "length", DescFunc: 0}, {Type: wasm.ExternTypeFunc, Module: "extism:host/env", Name: "output_set", DescFunc: 1}},
		FunctionSection: []wasm.Index{2},
		ExportSection:   []*wasm.Export{{Name: "http_invalid_pointer_v1", Type: wasm.ExternTypeFunc, Index: 3}},
		CodeSection:     []*wasm.Code{{LocalTypes: []wasm.ValueType{wasm.ValueTypeI64}, Body: []byte{wasm.OpcodeI64Const, 0x7f, wasm.OpcodeCall, 0, wasm.OpcodeLocalTee, 0, wasm.OpcodeLocalGet, 0, wasm.OpcodeCall, 1, wasm.OpcodeCall, 2, wasm.OpcodeI32Const, 0, wasm.OpcodeEnd}}},
	})
}
