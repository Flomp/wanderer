package pluginsystem

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"log"
	"log/slog"
	"strconv"
)

// diagnosticFailure labels a trusted local boundary. The underlying error is
// retained for errors.Is/As, but its text must never reach logs or public errors.
type diagnosticFailure struct {
	err   error
	cause string
}

func (e diagnosticFailure) Error() string { return PublicPluginError(e.err).Message }
func (e diagnosticFailure) Unwrap() error { return e.err }

// WithDiagnosticCause attaches a closed-set, bounded cause, never an arbitrary
// database, provider, URL, body, credential or panic string. Public error codes
// and status classification are unchanged.
func WithDiagnosticCause(err error, cause string) error {
	if err == nil {
		return nil
	}
	var existing diagnosticFailure
	if errors.As(err, &existing) {
		cause = existing.cause
	}
	return diagnosticFailure{err: err, cause: safeDiagnosticCause(cause)}
}

// The driver's numeric result code is safe; its message may contain SQL and
// parameters. Interpret codes only at a known local database boundary.
func diagnosticDetail(err error, cause string) string {
	switch cause {
	case "database_lookup", "database_save", "auto_merge", "trail_import":
	default:
		return ""
	}
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "not_found"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, context.Canceled):
		return "canceled"
	}
	var resultCode interface{ Code() int }
	if errors.As(err, &resultCode) {
		switch resultCode.Code() {
		case 2067:
			return "unique_constraint"
		case 1555:
			return "primary_key_constraint"
		case 787:
			return "foreign_key_constraint"
		case 1299:
			return "not_null_constraint"
		case 275:
			return "check_constraint"
		}
		switch resultCode.Code() & 0xff {
		case 5:
			return "database_busy"
		case 6:
			return "database_locked"
		case 8:
			return "database_read_only"
		case 10:
			return "database_io_error"
		case 11:
			return "database_corrupt"
		case 13:
			return "database_full"
		case 19:
			return "constraint_failed"
		}
	}
	return ""
}

func safeDiagnosticCause(cause string) string {
	switch cause {
	case "database_lookup", "database_save", "auto_merge", "trail_import", "batch_limit",
		"input_encoding", "file_read_failed", "actor_missing", "credentials_unreadable", "auth_context_invalid",
		"config_rejected", "config_shape_invalid", "required_config_missing", "config_value_type_invalid", "config_option_invalid",
		"connector_resolution_failed", "tls_mode_unsupported", "tls_custom_not_allowed", "tls_ca_invalid", "storage_origin_invalid",
		"request_policy_rejected", "host_auth_invalid", "request_body_invalid", "upload_policy_rejected",
		"request_build_failed", "connector_client_invalid", "network_unavailable", "response_policy_rejected",
		"response_body_incomplete", "response_body_oversized", "runtime_panic", "token_response_invalid", "media_fetch_failed", "photo_source_invalid":
		return cause
	default:
		return "local_failure"
	}
}

func diagnosticClass(code string) string {
	switch code {
	case "auth_failed", "invalid_grant", "invalid_client", "unauthorized_client", "credentials_unreadable":
		return "authentication"
	case "invalid_configuration", "invalid_request", "invalid_scope", "unsupported_grant_type", "unsupported_protocol":
		return "configuration"
	case "rate_limited":
		return "rate_limit"
	case "provider_unavailable", "temporary_unavailable":
		return "transient"
	case "unauthorized", "access_denied":
		return "authorization"
	case "not_found", "not_importable":
		return "item"
	case "runtime_error", "runtime_timeout", "plugin_error", "invalid_output":
		return "runtime"
	default:
		return "internal"
	}
}

// PluginDiagnosticAttrs is safe even for an arbitrary wrapped guest error.
// Only canonical error metadata and host-selected cause identifiers are used.
func PluginDiagnosticAttrs(err error) []any {
	e := PublicPluginError(err)
	attrs := []any{"code", e.Code, "class", diagnosticClass(e.Code), "operation", e.Operation, "http_status", e.HTTPStatus}
	if e.RetryAfterSeconds != nil {
		attrs = append(attrs, "retry_after_seconds", *e.RetryAfterSeconds)
	}
	var diagnostic diagnosticFailure
	if errors.As(err, &diagnostic) {
		cause := safeDiagnosticCause(diagnostic.cause)
		attrs = append(attrs, "cause", cause)
		if detail := diagnosticDetail(err, cause); detail != "" {
			attrs = append(attrs, "cause_detail", detail)
		}
	}
	return attrs
}

// LogPluginFailure's message and identifying attrs are host-owned. Never pass
// a provider URL/body or an arbitrary err.Error() in either of those arguments.
func LogPluginFailure(logger *slog.Logger, message string, err error, attrs ...any) {
	logger.Warn(message, append(attrs, PluginDiagnosticAttrs(err)...)...)
}

func safeGuestLogEvent(event string) string {
	switch event {
	case "request_started", "request_completed", "request_failed", "sync_started", "sync_completed", "sync_failed":
		return event
	default:
		return "guest_message_redacted"
	}
}

// Legacy guest text cannot be safely distinguished from provider response text.
// Structured events preserve diagnostics without allowing text to cross that
// boundary. Normalize again in the parent: the worker is not a trust boundary.
func logGuestEntry(entry HostLogEntry, sessionID string) {
	event := safeGuestLogEvent(entry.Event)
	if entry.Error != nil {
		e := NormalizePluginError(*entry.Error)
		retry := ""
		if e.RetryAfterSeconds != nil {
			retry = " retry_after_seconds=" + strconv.Itoa(*e.RetryAfterSeconds)
		}
		log.Printf("plugin log [%s]: session %s: event=%s code=%s class=%s operation=%s http_status=%d%s", entry.Level, sessionID, event, e.Code, diagnosticClass(e.Code), e.Operation, e.HTTPStatus, retry)
		return
	}
	log.Printf("plugin log [%s]: session %s: event=%s", entry.Level, sessionID, event)
}

// SourceDiagnosticID correlates guest sources without logging their free-text
// provider/IDs, which may carry URLs or credentials. It is not a record ID.
func SourceDiagnosticID(provider, externalID string) string {
	hash := sha256.New()
	var length [8]byte
	binary.LittleEndian.PutUint64(length[:], uint64(len(provider)))
	_, _ = hash.Write(length[:])
	_, _ = hash.Write([]byte(provider))
	_, _ = hash.Write([]byte(externalID))
	return hex.EncodeToString(hash.Sum(nil)[:8])
}
