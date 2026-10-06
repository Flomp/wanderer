package sdk

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// PluginFailure carries public error metadata without a provider response body,
// request URL, or arbitrary diagnostic text. It survives errors wrapped with %w.
type PluginFailure struct {
	PluginError
}

func (e PluginFailure) Error() string {
	return NormalizePluginError(e.PluginError).Code
}

// NormalizePluginError limits the wire error to known codes and safe metadata.
// Message is deliberately generated from Code, never copied from a provider.
func NormalizePluginError(e PluginError) PluginError {
	e.Code = strings.TrimSpace(e.Code)
	switch e.Code {
	case "auth_failed", "credentials_unreadable", "invalid_configuration", "invalid_grant", "unauthorized", "invalid_client", "unauthorized_client",
		"invalid_request", "invalid_scope", "unsupported_grant_type", "access_denied",
		"rate_limited", "provider_unavailable", "temporary_unavailable", "runtime_error",
		"runtime_timeout", "invalid_output", "internal_error", "plugin_error", "not_importable",
		"unsupported_protocol", "not_found":
	default:
		e.Code = "plugin_error"
	}
	e.Message = publicErrorMessage(e.Code)
	switch e.Operation {
	case "login", "oauth_exchange", "oauth_refresh", "list", "detail", "send", "host_http", "runtime":
	default:
		e.Operation = ""
	}
	if e.HTTPStatus < 100 || e.HTTPStatus > 599 {
		e.HTTPStatus = 0
	}
	if e.RetryAfterSeconds != nil {
		seconds := *e.RetryAfterSeconds
		if seconds < 0 {
			e.RetryAfterSeconds = nil
		} else {
			if seconds > 86400 {
				seconds = 86400
			}
			e.RetryAfterSeconds = &seconds
		}
	}
	return e
}

// SafeError preserves typed errors while discarding untrusted error strings.
// The caller supplies the operation because host transport errors are shared by
// login, list, detail and send. Unknown guest failures indicate invalid output.
func SafeError(err error, operation string) PluginError {
	e := PluginError{Code: "invalid_output"}
	var failure PluginFailure
	var failurePtr *PluginFailure
	if errors.As(err, &failure) {
		e = failure.PluginError
	} else if errors.As(err, &failurePtr) && failurePtr != nil {
		e = failurePtr.PluginError
	}
	if operation != "" {
		e.Operation = operation
	}
	return NormalizePluginError(e)
}

// ClassifyProviderError applies operation-specific HTTP fallbacks. The optional
// RFC error name must come from an OAuth error field, not response-description
// text. No provider body is accepted or retained here.
func ClassifyProviderError(operation string, status int, retryAfter, rfcError string, now time.Time) PluginError {
	rfcError = strings.TrimSpace(rfcError)
	e := PluginError{Operation: operation, HTTPStatus: status, RetryAfterSeconds: ParseRetryAfter(retryAfter, now)}
	if (operation == "oauth_refresh" || operation == "oauth_exchange") && rfcError != "" {
		switch rfcError {
		case "invalid_grant", "invalid_client", "unauthorized_client", "invalid_request", "invalid_scope", "unsupported_grant_type", "access_denied":
			e.Code = rfcError
		case "server_error", "temporarily_unavailable":
			e.Code = "provider_unavailable"
		default:
			e.Code = "invalid_request"
		}
	} else {
		switch {
		case status == 429:
			e.Code = "rate_limited"
		case status == 0 || status >= 500:
			e.Code = "provider_unavailable"
		case (operation == "oauth_refresh" || operation == "oauth_exchange") && status == 401:
			e.Code = "invalid_client"
		case operation == "oauth_refresh" && (status == 400 || status == 403):
			e.Code = "invalid_grant"
		case operation == "login" && (status == 400 || status == 401 || status == 403):
			e.Code = "auth_failed"
		case status == 401:
			e.Code = "unauthorized"
		case status == 403:
			e.Code = "access_denied"
		default:
			e.Code = "invalid_request"
		}
	}
	return NormalizePluginError(e)
}

// ResponseError converts an unsuccessful HTTP response to a typed safe error.
func ResponseError(operation string, response HostResponse) error {
	if response.Error != nil {
		e := *response.Error
		e.Operation = operation
		return PluginFailure{NormalizePluginError(e)}
	}
	return PluginFailure{ClassifyProviderError(operation, response.Status, response.FirstHeader("Retry-After"), "", time.Now())}
}

// ParseRetryAfter accepts delta-seconds and HTTP dates, including an immediate
// zero-second retry. Negative/malformed values are absent; valid delays are
// capped at 24 hours. The host chooses the fallback when the header is absent.
func ParseRetryAfter(value string, now time.Time) *int {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		if seconds < 0 {
			return nil
		}
		if seconds > 86400 {
			seconds = 86400
		}
		result := int(seconds)
		return &result
	}
	for _, layout := range []string{time.RFC1123, time.RFC1123Z, time.RFC850, time.ANSIC} {
		date, err := time.Parse(layout, value)
		if err != nil {
			continue
		}
		delay := date.Sub(now)
		seconds := int64(0)
		if delay > 0 {
			seconds = int64(delay / time.Second)
			if delay%time.Second != 0 {
				seconds++
			}
		}
		if seconds > 86400 {
			seconds = 86400
		}
		result := int(seconds)
		return &result
	}
	return nil
}

// publicErrorMessage provides a bounded English fallback for v1 hosts whose UI
// displays unknown-code messages. Provider diagnostics never enter this table.
func publicErrorMessage(code string) string {
	switch code {
	case "auth_failed":
		return "The provider rejected the login credentials."
	case "credentials_unreadable":
		return "Stored credentials could not be read. Enter them again or reconnect the plugin."
	case "invalid_configuration":
		return "The plugin configuration is invalid. Check the plugin settings."
	case "invalid_grant":
		return "Plugin authorization expired. Reconnect the plugin."
	case "unauthorized":
		return "The provider rejected authorization for this operation."
	case "invalid_client":
		return "The OAuth client credentials were rejected. Check the client ID and secret."
	case "unauthorized_client":
		return "The OAuth client credentials were rejected. Check the client ID and secret."
	case "invalid_request":
		return "The plugin sent an invalid request."
	case "invalid_scope":
		return "The provider rejected the plugin configuration. Check the app settings and permissions."
	case "unsupported_grant_type":
		return "The provider rejected the plugin configuration. Check the app settings and permissions."
	case "access_denied":
		return "The provider denied access to this item or operation."
	case "rate_limited":
		return "The provider rate limit was reached. Try again later."
	case "provider_unavailable":
		return "The provider is currently unavailable. Try again later."
	case "temporary_unavailable":
		return "The provider is currently unavailable. Try again later."
	case "runtime_error":
		return "The plugin failed unexpectedly."
	case "runtime_timeout":
		return "The plugin operation timed out."
	case "invalid_output":
		return "The plugin returned an invalid response."
	case "internal_error":
		return "The plugin failed unexpectedly."
	case "plugin_error":
		return "The plugin failed unexpectedly."
	case "not_importable":
		return "This activity has no importable track."
	case "unsupported_protocol":
		return "The plugin protocol is not supported."
	case "not_found":
		return "The requested item was not found."
	default:
		return "The plugin failed unexpectedly."
	}
}
