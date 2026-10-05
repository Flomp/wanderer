package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestProviderErrorsKeepOperationSpecificMeaning(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		operation string
		status    int
		rfc, want string
	}{
		{"login", 400, "", "auth_failed"},
		{"login", 403, "", "auth_failed"},
		{"login", 503, "", "provider_unavailable"},
		{"oauth_refresh", 400, "", "invalid_grant"},
		{"oauth_refresh", 403, "", "invalid_grant"},
		{"oauth_exchange", 400, "", "invalid_request"},
		{"oauth_exchange", 403, "", "access_denied"},
		{"oauth_refresh", 401, "", "invalid_client"},
		{"oauth_exchange", 401, "", "invalid_client"},
		{"oauth_refresh", 400, "invalid_scope", "invalid_scope"},
		{"oauth_refresh", 400, " invalid_grant ", "invalid_grant"},
		{"oauth_refresh", 403, "invalid_client", "invalid_client"},
		{"oauth_refresh", 400, "unrecognized-provider-text", "invalid_request"},
		{"oauth_exchange", 400, "server_error", "provider_unavailable"},
		{"oauth_refresh", 400, "temporarily_unavailable", "provider_unavailable"},
		{"detail", 403, "invalid_grant", "access_denied"},
		{"detail", 403, "", "access_denied"},
		{"send", 401, "", "unauthorized"},
		{"list", 429, "", "rate_limited"},
		{"host_http", 0, "", "provider_unavailable"},
	} {
		t.Run(fmt.Sprintf("%s-%d-%s", tc.operation, tc.status, tc.rfc), func(t *testing.T) {
			e := ClassifyProviderError(tc.operation, tc.status, "", tc.rfc, now)
			if e.Code != tc.want || e.Message != tc.want || e.Operation != tc.operation || (tc.status > 0 && e.HTTPStatus != tc.status) {
				t.Fatalf("unexpected public error: %+v", e)
			}
		})
	}
}

func TestWrappedHostErrorRetainsMetadataWithoutDiagnostics(t *testing.T) {
	secret := "https://token-secret@example.test/private?access_token=secret"
	seconds := 17
	failure := PluginFailure{PluginError{
		Code: "rate_limited", Message: secret, Operation: "host_http",
		HTTPStatus: 429, RetryAfterSeconds: &seconds,
	}}
	err := fmt.Errorf("provider replied %s: %w", secret, failure)
	safe := SafeError(err, "login")
	if safe.Code != "rate_limited" || safe.Operation != "login" || safe.HTTPStatus != 429 || safe.RetryAfterSeconds == nil || *safe.RetryAfterSeconds != 17 {
		t.Fatalf("wrapped transport metadata lost: %+v", safe)
	}
	data, err := json.Marshal(safe)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "secret") || strings.Contains(failure.Error(), "secret") {
		t.Fatalf("unsafe diagnostic escaped: %s", data)
	}
	for _, err := range []error{errors.New(secret), &PluginFailure{PluginError{Code: secret, Message: secret, Operation: secret, HTTPStatus: -1}}} {
		e := SafeError(err, "detail")
		data, _ := json.Marshal(e)
		if strings.Contains(string(data), "secret") || e.HTTPStatus != 0 || e.Operation != "detail" {
			t.Fatalf("untrusted error metadata escaped: %s", data)
		}
	}
}

func TestRetryAfterPreservesImmediateDateAndBoundedDelays(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		header string
		want   int
	}{
		{"0", 0}, {"17", 17}, {"172800", 86400},
		{now.Add(19 * time.Second).Format(time.RFC1123), 19},
		{now.Add(-time.Second).Format(time.RFC1123), 0},
	} {
		got := ParseRetryAfter(tc.header, now)
		if got == nil || *got != tc.want {
			t.Fatalf("Retry-After %q: got %v, want %d", tc.header, got, tc.want)
		}
	}
	for _, header := range []string{"", "-1", "not-a-header"} {
		if got := ParseRetryAfter(header, now); got != nil {
			t.Fatalf("invalid Retry-After %q produced %d", header, *got)
		}
	}
}

func TestResponseErrorPreservesLegacyWireAndRetryHeader(t *testing.T) {
	var response HostResponse
	if err := json.Unmarshal([]byte(`{"status":0,"error":{"code":"provider_unavailable","message":"sensitive URL or body"}}`), &response); err != nil {
		t.Fatal(err)
	}
	e := SafeError(ResponseError("login", response), "login")
	if e.Code != "provider_unavailable" || e.Message != "provider_unavailable" || e.Operation != "login" {
		t.Fatalf("legacy error was not preserved safely: %+v", e)
	}
	response = HostResponse{Status: 429, HeaderValues: map[string][]string{"retry-after": {"37"}}, BodyBase64: "sensitive-body"}
	e = SafeError(ResponseError("list", response), "list")
	if e.Code != "rate_limited" || e.HTTPStatus != 429 || e.RetryAfterSeconds == nil || *e.RetryAfterSeconds != 37 {
		t.Fatalf("HTTP response metadata was not preserved: %+v", e)
	}
}
