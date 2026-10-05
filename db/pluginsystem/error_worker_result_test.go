package pluginsystem

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExtismResultPreservesStructuredGuestFailure(t *testing.T) {
	raw := `{"code":"rate_limited","message":"provider-body-secret","operation":"list","httpStatus":429,"retryAfterSeconds":120}`
	for _, callErr := range []error{nil, errors.New(raw)} {
		got := pluginErrorForResult("list_routes_v1", 1, raw, callErr)
		if got == nil || got.Code != "rate_limited" || got.Message != "rate_limited" || got.Operation != OperationList || got.HTTPStatus != 429 || got.RetryAfterSeconds == nil || *got.RetryAfterSeconds != 120 {
			t.Fatalf("guest failure became runtime error: %#v", got)
		}
	}
	if got := pluginErrorForResult("list_routes_v1", 1, "panic-secret-value", errors.New("panic-secret-value")); got != nil {
		t.Fatalf("trap treated as guest rejection: %#v", got)
	}
	if got := pluginErrorForResult("list_routes_v1", 0, raw, errors.New("trap")); got != nil {
		t.Fatalf("zero code treated as guest rejection: %#v", got)
	}
	if got := pluginErrorForResult("list_routes_v1", 1, "", nil); got == nil || got.Code != "plugin_error" {
		t.Fatalf("missing old guest error fallback: %#v", got)
	}
}

func TestWorkerSessionStartupDeadlineIsTimeout(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	_, err := (WorkerRuntime{Executable: "must-not-launch-secret-path"}).OpenSession(ctx, LocalPlugin{}, RequestPolicyContext{})
	if err == nil {
		t.Fatal("expected expired startup deadline")
	}
	failure := PublicPluginError(err)
	if failure.Code != "runtime_timeout" || failure.Operation != OperationRuntime || err.Error() != "runtime_timeout" {
		t.Fatalf("startup deadline became %#v / %v", failure, err)
	}
}
