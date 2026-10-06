package pluginsystem

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pocketbase/util"

	extism "github.com/extism/go-sdk"
)

// These calls exercise the compiled TinyGo HTTP wrappers and export errors.
// Ordinary Go tests cannot see those files. CI supplies the bundle directory
// after its existing build; no provider account or network interface is needed.
func TestBundledPluginErrors(t *testing.T) {
	bundleRoot := os.Getenv("WANDERER_PLUGIN_BUNDLE_DIR")
	if bundleRoot == "" {
		t.Skip("run after make plugins-build with WANDERER_PLUGIN_BUNDLE_DIR=/path/to/wanderer/plugins")
	}
	const privateDetail = "sentinel-provider-body token=private-token https://private.example/?secret=hidden"
	cases := []struct {
		name, plugin, export, code, operation string
		status                                int
		retryAfter                            string
		retrySeconds                          *int
		transportError                        bool
	}{
		{name: "dashboard unavailable", plugin: "hammerhead", export: "refresh_session_v1", status: 503, code: "provider_unavailable", operation: OperationLogin},
		{name: "dashboard invalid credentials", plugin: "hammerhead", export: "refresh_session_v1", status: 403, code: "auth_failed", operation: OperationLogin},
		{name: "dashboard rate limited", plugin: "komoot", export: "refresh_session_v1", status: 429, retryAfter: "37", retrySeconds: bundledRetrySeconds(37), code: "rate_limited", operation: OperationLogin},
		{name: "immediate retry", plugin: "komoot", export: "refresh_session_v1", status: 429, retryAfter: "0", retrySeconds: bundledRetrySeconds(0), code: "rate_limited", operation: OperationLogin},
		{name: "list forbidden", plugin: "strava", export: "list_routes_v1", status: 403, code: "access_denied", operation: OperationList},
		{name: "list rate limited", plugin: "strava", export: "list_activities_v1", status: 429, retryAfter: "23", retrySeconds: bundledRetrySeconds(23), code: "rate_limited", operation: OperationList},
		{name: "transport failure", plugin: "hammerhead", export: "refresh_session_v1", code: "provider_unavailable", operation: OperationLogin, transportError: true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(bundleRoot, tt.plugin, "dist", tt.plugin)
			manifestJSON, err := os.ReadFile(filepath.Join(dir, "plugin.json"))
			if err != nil {
				t.Fatal(err)
			}
			var manifest Manifest
			if err := json.Unmarshal(manifestJSON, &manifest); err != nil {
				t.Fatal(err)
			}
			if err := ValidateManifest(manifest); err != nil {
				t.Fatal(err)
			}
			policy := RequestPolicyContext{Connectors: map[string]ResolvedConnectorTarget{}}
			for _, connector := range manifest.Permissions.Network.Connectors {
				base, err := url.Parse(connector.FixedBaseURL)
				if err != nil || base.Host == "" {
					t.Fatalf("invalid public connector %s", connector.Name)
				}
				policy.Connectors[connector.Name] = ResolvedConnectorTarget{
					Name: connector.Name, Type: connector.Type, BaseURL: connector.FixedBaseURL,
					BasePath: base.Path, TLS: ConnectorTLSConfig{Mode: TLSModeSystem},
				}
			}
			original := newConnectorHTTPClient
			requests := 0
			newConnectorHTTPClient = func(_ util.ConnectorHTTPPolicy, _ func(*http.Request, []*http.Request) error) (*http.Client, error) {
				return &http.Client{Transport: bundledErrorTransport(func(req *http.Request) (*http.Response, error) {
					requests++
					if tt.transportError {
						return nil, errors.New(privateDetail)
					}
					return &http.Response{
						StatusCode: tt.status, Header: http.Header{"Content-Type": {"text/plain"}, "Retry-After": {tt.retryAfter}},
						Body: io.NopCloser(strings.NewReader(privateDetail)), Request: req,
					}, nil
				})}, nil
			}
			t.Cleanup(func() { newConnectorHTTPClient = original })
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			instance, err := extism.NewPlugin(ctx, extism.Manifest{Wasm: []extism.Wasm{
				extism.WasmFile{Path: filepath.Join(dir, manifest.Runtime.Entrypoint)},
			}}, extism.PluginConfig{EnableWasi: true}, extismHostFunctions(manifest, policy))
			if err != nil {
				t.Fatal(err)
			}
			defer instance.Close(ctx)
			code, _, err := instance.CallWithContext(ctx, tt.export, []byte(`{"auth":{"email":"user@example.com","password":"password-secret","accessToken":"access-secret"},"config":{"plugin":{}}}`))
			if code == 0 {
				t.Fatalf("expected guest failure, code=%d error=%v", code, err)
			}
			if requests != 1 {
				t.Fatalf("expected one intercepted provider request, got %d", requests)
			}
			raw := instance.GetErrorWithContext(ctx)
			if strings.Contains(raw, "sentinel") || strings.Contains(raw, "secret") || strings.Contains(raw, "private.example") {
				t.Fatalf("guest returned sensitive diagnostics: %q", raw)
			}
			// The v0.21.x UI shows message for codes it does not know. Inspect
			// the actual compiled guest wire before modern-host normalization.
			var guest PluginError
			if err := json.Unmarshal([]byte(raw), &guest); err != nil {
				t.Fatal(err)
			}
			if guest.Code != tt.code || guest.Message == guest.Code || len(guest.Message) > 200 || !strings.Contains(guest.Message, " ") {
				t.Fatalf("unreadable legacy-host fallback: %s", raw)
			}
			failure := pluginErrorForCode(tt.export, code, raw)
			if failure.Code != tt.code || failure.Operation != tt.operation || failure.HTTPStatus != tt.status || failure.Message != tt.code {
				t.Fatalf("unexpected public error: %+v", failure)
			}
			if (failure.RetryAfterSeconds == nil) != (tt.retrySeconds == nil) || (tt.retrySeconds != nil && *failure.RetryAfterSeconds != *tt.retrySeconds) {
				t.Fatalf("unexpected retry delay: %+v", failure)
			}
			// Extism returns both a nonzero code and an error string for SetError.
			// The real worker must forward that guest failure rather than treating
			// the accompanying Go error as a fatal runtime trap.
			requests = 0
			workerFailure := bundledWorkerFailure(t, ctx, manifest, policy, filepath.Join(dir, manifest.Runtime.Entrypoint), tt.export)
			if workerFailure.Code != failure.Code || workerFailure.Message != failure.Message || workerFailure.Operation != failure.Operation || workerFailure.HTTPStatus != failure.HTTPStatus {
				t.Fatalf("worker changed the guest failure: %+v", workerFailure)
			}
			if (workerFailure.RetryAfterSeconds == nil) != (failure.RetryAfterSeconds == nil) || (failure.RetryAfterSeconds != nil && *workerFailure.RetryAfterSeconds != *failure.RetryAfterSeconds) {
				t.Fatalf("worker lost the retry delay: %+v", workerFailure)
			}
			if requests != 1 {
				t.Fatalf("worker made %d intercepted provider requests", requests)
			}
		})
	}
}

type bundledErrorTransport func(*http.Request) (*http.Response, error)

func (f bundledErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func bundledRetrySeconds(value int) *int { return &value }

func bundledWorkerFailure(t *testing.T, ctx context.Context, manifest Manifest, policy RequestPolicyContext, wasmPath, export string) PluginError {
	t.Helper()
	workerInput, parentOutput := io.Pipe()
	parentInput, workerOutput := io.Pipe()
	closePipes := func() {
		_ = parentOutput.Close()
		_ = parentInput.Close()
		_ = workerInput.Close()
		_ = workerOutput.Close()
	}
	stop := context.AfterFunc(ctx, closePipes)
	done := make(chan int, 1)
	go func() {
		done <- RunPluginWorker(ctx, workerInput, workerOutput, io.Discard)
	}()
	defer func() {
		closePipes()
		stop()
		select {
		case <-done:
		case <-ctx.Done():
			t.Error("worker did not stop")
		}
	}()
	message, err := workerMessageWithData(workerMessageCallExport, workerCallExport{
		WASMPath: wasmPath, Export: export, SessionID: "bundled-errors",
		InputBase64: encodeWorkerBytes([]byte(`{"auth":{"email":"user@example.com","password":"password-secret","accessToken":"access-secret"},"config":{"plugin":{}}}`)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := writeWorkerMessage(parentOutput, defaultWorkerRequestMaxBytes, message); err != nil {
		t.Fatal(err)
	}
	parent := workerRuntimeSession{
		plugin: LocalPlugin{Manifest: manifest}, policy: policy,
		stdin: parentOutput, responseMaxBytes: defaultWorkerResponseMaxBytes,
	}
	for {
		reply, err := readWorkerMessage(parentInput, defaultWorkerResponseMaxBytes)
		if err != nil {
			t.Fatal(err)
		}
		switch reply.Type {
		case workerMessageHostHTTPRequest:
			if err := parent.handleHostHTTPRequest(ctx, reply); err != nil {
				t.Fatal(err)
			}
		case workerMessageCallResult:
			result, err := workerData[workerCallResult](reply)
			if err != nil || result.PluginError == nil {
				t.Fatalf("worker did not return a typed guest failure: %+v, %v", result, err)
			}
			return *result.PluginError
		default:
			t.Fatalf("worker returned %s instead of a guest failure", reply.Type)
		}
	}
}
