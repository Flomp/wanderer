package pluginsystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestProviderErrorClassificationIsOperationSpecific(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	for _, operation := range []string{OperationLogin, OperationOAuthExchange, OperationOAuthRefresh, OperationList, OperationDetail, OperationSend} {
		for _, status := range []int{400, 401, 403, 429, 503} {
			code := "invalid_request"
			switch {
			case status == 429:
				code = "rate_limited"
			case status == 503:
				code = "provider_unavailable"
			case (operation == OperationOAuthExchange || operation == OperationOAuthRefresh) && status == 401:
				code = "invalid_client"
			case operation == OperationOAuthRefresh && (status == 400 || status == 403):
				code = "invalid_grant"
			case operation == OperationLogin:
				code = "auth_failed"
			case status == 401:
				code = "unauthorized"
			case status == 403:
				code = "access_denied"
			}
			t.Run(fmt.Sprintf("%s/%d", operation, status), func(t *testing.T) {
				got := ClassifyProviderError(operation, status, "120", "", now)
				if got.Code != code || got.Operation != operation || got.HTTPStatus != status || got.Message != code {
					t.Fatalf("classification = %#v, want code %s", got, code)
				}
				if got.RetryAfterSeconds == nil || *got.RetryAfterSeconds != 120 {
					t.Fatalf("retry metadata = %#v", got)
				}
			})
		}
	}
	for _, code := range []string{"invalid_request", "invalid_scope", "unsupported_grant_type", "invalid_client", "unauthorized_client", "invalid_grant"} {
		got := ClassifyProviderError(OperationOAuthRefresh, 400, "", code, now)
		if got.Code != code {
			t.Fatalf("explicit RFC %s became %#v", code, got)
		}
	}
	if got := ClassifyProviderError(OperationOAuthRefresh, 403, "", "provider-private-secret", now); got.Code != "invalid_request" {
		t.Fatalf("unknown explicit error reached HTTP fallback: %#v", got)
	}
	if got := ClassifyProviderError(OperationList, 403, "", "invalid_grant", now); got.Code != "access_denied" {
		t.Fatalf("RFC identifier used outside token operation: %#v", got)
	}
}

func TestRetryAfterDatesZeroFallbackAndBounds(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 500000000, time.UTC)
	for _, tc := range []struct {
		header string
		want   *int
	}{
		{"", nil}, {"-1", nil}, {"nonsense", nil}, {"0", intPtr(0)},
		{"120", intPtr(120)}, {"999999999999", intPtr(86400)},
		{now.Add(3 * time.Second).Format(http.TimeFormat), intPtr(3)},
		{now.Add(-time.Hour).Format(http.TimeFormat), intPtr(0)},
	} {
		got := ParseRetryAfter(tc.header, now)
		if (got == nil) != (tc.want == nil) || got != nil && *got != *tc.want {
			t.Errorf("ParseRetryAfter(%q) = %v, want %v", tc.header, got, tc.want)
		}
	}
	zero := 0
	update := InstanceStatusForPluginError(PluginError{Code: "rate_limited", RetryAfterSeconds: &zero}, now)
	if update.RetryNotBefore == nil || !update.RetryNotBefore.Equal(now) {
		t.Fatalf("explicit zero replaced by fallback: %#v", update)
	}
	update = InstanceStatusForPluginError(PluginError{Code: "rate_limited"}, now)
	if update.RetryNotBefore == nil || !update.RetryNotBefore.Equal(now.Add(time.Hour)) {
		t.Fatalf("missing fallback: %#v", update)
	}
}

func intPtr(value int) *int { return &value }

func TestPublicPluginErrorsDropUntrustedMessages(t *testing.T) {
	secret := "provider-body-secret https://user:password@provider.test/path?token=secret"
	for _, tc := range []struct {
		err  error
		code string
	}{
		{errors.New(secret), "internal_error"},
		{PluginCapabilityError{Err: &PluginError{Code: "rate_limited", Message: secret}}, "rate_limited"},
		{PluginCallError{PluginError: PluginError{Code: "invalid_grant", Message: secret}}, "invalid_grant"},
		{RuntimeSessionFatalError{Err: errors.New(secret)}, "runtime_error"},
		{RuntimeSessionFatalError{Err: context.DeadlineExceeded}, "runtime_timeout"},
		{&url.Error{Op: "Get", URL: secret, Err: errors.New(secret)}, "provider_unavailable"},
		{PluginFailure{PluginError: PluginError{Code: "custom-secret", Message: secret, Operation: secret}}, "plugin_error"},
	} {
		got := PublicPluginError(tc.err)
		if got.Code != tc.code || got.Message != tc.code {
			t.Fatalf("safe error = %#v, want %s", got, tc.code)
		}
		encoded, _ := json.Marshal(InstanceStatusForError(tc.err, time.Now()))
		if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "https:") {
			t.Fatalf("diagnostic leak: %s", encoded)
		}
	}
}

func TestOAuthExchangeRejectsProviderBodiesWithoutLeaking(t *testing.T) {
	for _, tc := range []struct {
		status            int
		grant, body, want string
	}{
		{400, "refresh_token", `{"error":"invalid_grant","error_description":"body-secret"}`, "invalid_grant"},
		{400, "refresh_token", `{"error":"invalid_scope","error_description":"body-secret"}`, "invalid_scope"},
		{403, "authorization_code", `body-secret`, "access_denied"},
		{401, "refresh_token", `body-secret`, "invalid_client"},
		{429, "refresh_token", `body-secret`, "rate_limited"},
		{503, "refresh_token", `body-secret`, "provider_unavailable"},
		{200, "refresh_token", `{"access_token":123,"body":"body-secret"}`, "invalid_output"},
	} {
		t.Run(fmt.Sprintf("%s/%d/%s", tc.grant, tc.status, tc.want), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "120")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			manifest := testHostManifest(t, server.URL)
			_, err := ExchangeOAuthToken(context.Background(), manifest, AuthContext{TokenURL: server.URL + "/v1"},
				map[string]any{"clientId": "client-secret", "clientSecret": "client-secret"}, map[string]string{"grant_type": tc.grant, "refresh_token": "refresh-secret", "code": "code-secret"})
			if err == nil {
				t.Fatal("expected rejection")
			}
			safe := PublicPluginError(err)
			if safe.Code != tc.want {
				t.Fatalf("got %#v, want %s", safe, tc.want)
			}
			if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), server.URL) {
				t.Fatalf("leaked response/URL: %v", err)
			}
			if tc.status != 200 && (safe.RetryAfterSeconds == nil || *safe.RetryAfterSeconds != 120 || safe.HTTPStatus != tc.status) {
				t.Fatalf("lost HTTP metadata: %#v", safe)
			}
		})
	}
}

func TestHostHTTPFailurePreservesStatusInsteadOfMediaType(t *testing.T) {
	useUnsafeTestHTTPClient(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(503)
		_, _ = w.Write([]byte("proxy-body-secret"))
	}))
	defer server.Close()
	spec := HostRequestSpec{Method: "GET", Target: RequestTarget{Type: "connector", Connector: "api", Path: "/v1"}, Expect: ResponseExpect{ContentTypes: []string{"application/json"}}}
	data, _ := json.Marshal(spec)
	got := executeHostHTTPRequest(context.Background(), testHostManifest(t, server.URL), testHostPolicy(t, server.URL), data)
	if got.Error != nil || got.Status != 503 {
		t.Fatalf("HTTP failure hidden by media type: %#v", got)
	}
	server.Close()
	got = executeHostHTTPRequest(context.Background(), testHostManifest(t, server.URL), testHostPolicy(t, server.URL), data)
	if got.Error == nil || got.Error.Code != "provider_unavailable" {
		t.Fatalf("transport error = %#v", got)
	}
	encoded, _ := json.Marshal(got)
	if strings.Contains(string(encoded), server.URL) || strings.Contains(string(encoded), "secret") {
		t.Fatalf("transport leaked: %s", encoded)
	}
}

func TestOAuthLocalEndpointRejectionIsConfigurationFailure(t *testing.T) {
	for _, tokenURL := range []string{"https://[invalid", "file:///private", "https://not-in-manifest.example/token"} {
		for _, grantType := range []string{"authorization_code", "refresh_token"} {
			_, err := ExchangeOAuthToken(context.Background(), Manifest{}, AuthContext{TokenURL: tokenURL}, nil, map[string]string{"grant_type": grantType})
			got := PublicPluginError(err)
			if got.Code != "invalid_configuration" || got.HTTPStatus != 0 || strings.Contains(got.Message, tokenURL) {
				t.Fatalf("local token URL %q classified as %#v", tokenURL, got)
			}
		}
	}
}
