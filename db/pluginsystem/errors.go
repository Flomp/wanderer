package pluginsystem

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Operations describe the failed call, without carrying request URLs, bodies,
// account IDs or credentials. New fields are optional in the v1 wire format.
const (
	OperationLogin         = "login"
	OperationOAuthExchange = "oauth_exchange"
	OperationOAuthRefresh  = "oauth_refresh"
	OperationList          = "list"
	OperationDetail        = "detail"
	OperationSend          = "send"
	OperationHostHTTP      = "host_http"
	OperationRuntime       = "runtime"
)

const maxRetryAfterSeconds = 24 * 60 * 60

// PluginFailure is safe to return, persist and log. It intentionally does not
// retain a wrapped provider error: HTTP errors can include secret-bearing URLs.
type PluginFailure struct{ PluginError PluginError }

func (e PluginFailure) Error() string { return NormalizePluginError(e.PluginError).Message }

// NormalizePluginError is the trust boundary for guest and legacy errors.
// Codes remain compatible; arbitrary messages and unknown metadata never cross
// into instance diagnostics, API responses or application logs.
func NormalizePluginError(e PluginError) PluginError {
	switch strings.TrimSpace(e.Code) {
	case "auth_failed", "credentials_unreadable", "invalid_configuration", "invalid_grant", "unauthorized", "invalid_client", "unauthorized_client",
		"invalid_request", "invalid_scope", "unsupported_grant_type", "access_denied",
		"rate_limited", "provider_unavailable", "temporary_unavailable", "internal_error",
		"plugin_error", "runtime_error", "runtime_timeout", "invalid_output", "not_importable", "not_found", "unsupported_protocol":
		e.Code = strings.TrimSpace(e.Code)
	default:
		e.Code = "plugin_error"
	}
	e.Message = e.Code
	switch e.Operation {
	case OperationLogin, OperationOAuthExchange, OperationOAuthRefresh, OperationList,
		OperationDetail, OperationSend, OperationHostHTTP, OperationRuntime:
	default:
		e.Operation = ""
	}
	if e.HTTPStatus < 100 || e.HTTPStatus > 599 {
		e.HTTPStatus = 0
	}
	if e.RetryAfterSeconds != nil {
		value := *e.RetryAfterSeconds
		if value < 0 {
			e.RetryAfterSeconds = nil
		} else {
			if value > maxRetryAfterSeconds {
				value = maxRetryAfterSeconds
			}
			e.RetryAfterSeconds = &value
		}
	}
	return e
}

// PublicPluginError preserves typed safe failures, including legacy ABI codes,
// while treating untyped errors as local failures rather than guessing that the
// user's provider credentials are wrong.
func PublicPluginError(err error) PluginError {
	var failure PluginFailure
	var failurePtr *PluginFailure
	var capability PluginCapabilityError
	var call PluginCallError
	if errors.As(err, &failure) {
		return NormalizePluginError(failure.PluginError)
	}
	if errors.As(err, &failurePtr) && failurePtr != nil {
		return NormalizePluginError(failurePtr.PluginError)
	}
	if errors.As(err, &capability) && capability.Err != nil {
		return NormalizePluginError(*capability.Err)
	}
	if errors.As(err, &call) {
		return NormalizePluginError(call.PluginError)
	}
	code := "internal_error"
	operation := ""
	if IsRuntimeSessionFatalError(err) || errors.Is(err, ErrRuntimeUnavailable) {
		code, operation = "runtime_error", OperationRuntime
	}
	if errors.Is(err, context.DeadlineExceeded) {
		code, operation = "runtime_timeout", OperationRuntime
	}
	var network *url.Error
	if errors.As(err, &network) {
		code, operation = "provider_unavailable", OperationHostHTTP
	}
	return NormalizePluginError(PluginError{Code: code, Operation: operation})
}

func Failure(code, operation string) error {
	return PluginFailure{PluginError: NormalizePluginError(PluginError{Code: code, Operation: operation})}
}

// ClassifyProviderError never accepts a provider message or body. For OAuth the
// caller may pass only the RFC error identifier parsed from a bounded response.
// Explicit RFC failures take precedence over HTTP fallback heuristics.
func ClassifyProviderError(operation string, status int, retryAfter, rfcError string, now time.Time) PluginError {
	code := "invalid_request"
	if operation == OperationOAuthExchange || operation == OperationOAuthRefresh {
		switch strings.TrimSpace(rfcError) {
		case "invalid_grant", "invalid_client", "unauthorized_client", "invalid_request", "invalid_scope", "unsupported_grant_type", "access_denied":
			code = strings.TrimSpace(rfcError)
		case "temporarily_unavailable", "server_error":
			code = "provider_unavailable"
		case "":
			code = providerHTTPFallback(operation, status)
		default:
			// An unknown explicit identifier is a contract failure; don't turn
			// it into invalid_grant via the refresh-only 400/403 fallback.
			code = "invalid_request"
		}
	} else {
		code = providerHTTPFallback(operation, status)
	}
	return NormalizePluginError(PluginError{Code: code, Operation: operation, HTTPStatus: status,
		RetryAfterSeconds: ParseRetryAfter(retryAfter, now)})
}

func providerHTTPFallback(operation string, status int) string {
	switch {
	case status == http.StatusTooManyRequests:
		return "rate_limited"
	case status == 0 || status >= 500:
		return "provider_unavailable"
	case (operation == OperationOAuthExchange || operation == OperationOAuthRefresh) && status == http.StatusUnauthorized:
		return "invalid_client"
	case operation == OperationOAuthRefresh && (status == http.StatusBadRequest || status == http.StatusForbidden):
		return "invalid_grant"
	case operation == OperationLogin && (status == http.StatusBadRequest || status == http.StatusUnauthorized || status == http.StatusForbidden):
		return "auth_failed"
	case status == http.StatusUnauthorized:
		return "unauthorized"
	case status == http.StatusForbidden:
		return "access_denied"
	default:
		return "invalid_request"
	}
}

// Retry-After supports seconds and HTTP dates. Invalid headers use the caller's
// documented fallback (one hour for rate limits); valid zero/past dates do not.
// Clamp long waits to one day before conversion to a duration.
func ParseRetryAfter(header string, now time.Time) *int {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil
	}
	var seconds int64
	if parsed, err := strconv.ParseInt(header, 10, 64); err == nil {
		if parsed < 0 {
			return nil
		}
		seconds = parsed
	} else if date, err := http.ParseTime(header); err == nil {
		delay := date.Sub(now)
		seconds = 0
		if delay > 0 {
			seconds = int64(delay / time.Second)
			if delay%time.Second != 0 {
				seconds++
			}
		}
	} else {
		return nil
	}
	if seconds > maxRetryAfterSeconds {
		seconds = maxRetryAfterSeconds
	}
	value := int(seconds)
	return &value
}
