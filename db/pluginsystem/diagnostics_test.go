package pluginsystem

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"testing"
)

func capturedDiagnostic(t *testing.T, err error) (map[string]any, string) {
	t.Helper()
	var output bytes.Buffer
	LogPluginFailure(slog.New(slog.NewJSONHandler(&output, nil)), "sync failed", err, "instance", "safe-id")
	var record map[string]any
	if e := json.Unmarshal(output.Bytes(), &record); e != nil {
		t.Fatal(e)
	}
	return record, output.String()
}

func TestDiagnosticKeepsClassificationAndHTTPMetadataWithoutFreeText(t *testing.T) {
	retry := 120
	err := fmt.Errorf("outer-secret: %w", PluginFailure{PluginError: PluginError{
		Code: "rate_limited", Message: "provider-body-secret https://user:secret@example.test?token=secret",
		Operation: OperationList, HTTPStatus: 429, RetryAfterSeconds: &retry,
	}})
	record, output := capturedDiagnostic(t, err)
	for key, want := range map[string]any{"code": "rate_limited", "class": "rate_limit", "operation": OperationList, "http_status": float64(429), "retry_after_seconds": float64(120)} {
		if record[key] != want {
			t.Fatalf("diagnostic %s = %#v, want %#v: %s", key, record[key], want, output)
		}
	}
	if strings.Contains(output, "secret") || strings.Contains(output, "https://") {
		t.Fatalf("unsafe diagnostic: %s", output)
	}
}

func TestLocalDiagnosticCausesAreBoundedAndPreserveErrorIdentity(t *testing.T) {
	for _, tc := range []struct {
		cause, want string
		err         error
	}{
		{"database_save", "database_save", fmt.Errorf("secret SQL: %w", context.DeadlineExceeded)},
		{"auto_merge", "auto_merge", errors.New("secret source-file path")},
		{"connector_resolution_failed", "connector_resolution_failed", Failure("invalid_configuration", OperationHostHTTP)},
		{"secret unknown cause https://token.invalid", "local_failure", errors.New("secret database error")},
	} {
		err := WithDiagnosticCause(tc.err, tc.cause)
		if !errors.Is(err, tc.err) {
			t.Fatal("diagnostic wrapper lost error identity")
		}
		record, output := capturedDiagnostic(t, err)
		if record["cause"] != tc.want {
			t.Fatalf("lost boundary: %s", output)
		}
		if tc.cause == "database_save" && record["cause_detail"] != "deadline_exceeded" {
			t.Fatalf("lost safe DB cause: %s", output)
		}
		if strings.Contains(output, "secret") || strings.Contains(err.Error(), "secret") {
			t.Fatalf("leaked internal cause: %s / %v", output, err)
		}
	}
	specific := WithDiagnosticCause(Failure("invalid_configuration", OperationHostHTTP), "tls_mode_unsupported")
	record, _ := capturedDiagnostic(t, WithDiagnosticCause(specific, "config_rejected"))
	if record["cause"] != "tls_mode_unsupported" {
		t.Fatal("generic boundary overwrote specific cause")
	}
}

func TestDatabaseDiagnosticUsesRealDriverCodeAndHidesSQL(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE secret_table (value TEXT UNIQUE)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO secret_table VALUES (?)", "secret-value"); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO secret_table VALUES (?)", "secret-value")
	if err == nil {
		t.Fatal("expected real unique constraint")
	}
	record, output := capturedDiagnostic(t, WithDiagnosticCause(err, "database_save"))
	if record["cause_detail"] != "unique_constraint" {
		t.Fatalf("lost driver constraint: %s", output)
	}
	if strings.Contains(output, "secret") || strings.Contains(output, "INSERT") {
		t.Fatalf("leaked SQL/parameter: %s", output)
	}
}

func TestHostPolicyDiagnosticRecordsBoundaryWithoutGuestURL(t *testing.T) {
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	manifest := testHostManifest(t, "https://provider.test")
	policy := testHostPolicy(t, "https://provider.test")
	data := []byte(`{"method":"GET","target":{"type":"connector","connector":"secret-invalid","path":"/?token=secret"}}`)
	response := executeHostHTTPRequest(context.Background(), manifest, policy, data)
	if response.Error == nil {
		t.Fatal("invalid connector was accepted")
	}
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["cause"] != "request_policy_rejected" || record["operation"] != OperationHostHTTP || record["code"] != "invalid_request" {
		t.Fatalf("missing policy diagnostic: %s", output.String())
	}
	if strings.Contains(output.String(), "secret") || strings.Contains(output.String(), "provider.test") {
		t.Fatalf("leaked request: %s", output.String())
	}
}

func TestStructuredGuestEventSurvivesRPCAndNormalizesAgain(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })
	entry, err := parseHostLogEntry([]byte(`{"level":"warn","message":"provider-body-secret","event":"request_failed","error":{"code":"rate_limited","message":"https://secret.invalid","operation":"detail","httpStatus":429}}`))
	if err != nil {
		t.Fatal(err)
	}
	msg, err := workerMessageWithData(workerMessageHostLog, workerHostLog{Level: entry.Level, Message: entry.Message, Event: entry.Event, Error: entry.Error, SessionID: "secret-forged-session"})
	if err != nil {
		t.Fatal(err)
	}
	session := workerRuntimeSession{sessionID: "trusted-session"}
	session.handleHostLog(msg)
	got := output.String()
	for _, want := range []string{"trusted-session", "event=request_failed", "code=rate_limited", "class=rate_limit", "operation=detail", "http_status=429"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing guest diagnostic %s: %s", want, got)
		}
	}
	if strings.Contains(got, "secret") || strings.Contains(got, "https://") {
		t.Fatalf("unsafe guest event: %s", got)
	}
	output.Reset()
	msg, err = workerMessageWithData(workerMessageHostLog, workerHostLog{Level: "error", Message: "secret", Event: "secret-event", Error: &PluginError{Code: "secret-code", Operation: "secret-operation", HTTPStatus: 999}})
	if err != nil {
		t.Fatal(err)
	}
	session.handleHostLog(msg)
	got = output.String()
	if !strings.Contains(got, "event=guest_message_redacted") || !strings.Contains(got, "code=plugin_error") || strings.Contains(got, "secret") {
		t.Fatalf("parent accepted arbitrary worker data: %s", got)
	}
}

func TestNestedDiagnosticWrapperHidesOuterErrorText(t *testing.T) {
	original := WithDiagnosticCause(Failure("provider_unavailable", OperationHostHTTP), "network_unavailable")
	outer := fmt.Errorf("URL-secret: %w", original)
	err := WithDiagnosticCause(outer, "trail_import")
	record, output := capturedDiagnostic(t, err)
	if strings.Contains(err.Error(), "secret") || strings.Contains(output, "secret") || record["cause"] != "network_unavailable" {
		t.Fatalf("nested wrapper lost privacy/specificity: %s / %v", output, err)
	}
}

func TestSourceDiagnosticIDIsBoundedAndAvoidsAmbiguousParts(t *testing.T) {
	value := SourceDiagnosticID("provider-secret", "https://token-secret.invalid")
	if len(value) != 16 || strings.Contains(value, "secret") {
		t.Fatalf("unsafe source correlation: %q", value)
	}
	if SourceDiagnosticID("a\x00b", "c") == SourceDiagnosticID("a", "b\x00c") {
		t.Fatal("source boundaries collide before hashing")
	}
}
