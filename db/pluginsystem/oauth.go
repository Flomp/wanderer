package pluginsystem

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

const (
	AuthFieldOAuthContext = "oauthContext"
	AuthFieldTokenType    = "tokenType"
	AuthFieldExpiresAt    = "expiresAt"
	AuthFieldScope        = "scope"
)

type OAuthTokenResponse struct {
	AccessToken    string          `json:"access_token"`
	RefreshToken   string          `json:"refresh_token,omitempty"`
	TokenType      string          `json:"token_type,omitempty"`
	ExpiresIn      int             `json:"expires_in,omitempty"`
	Scope          string          `json:"scope,omitempty"`
	Raw            json.RawMessage `json:"-"`
	ProviderUserID string          `json:"providerUserId,omitempty"`
}

// OAuthContext uses the same unambiguous selection as other provider operations.
func OAuthContext(plugin LocalPlugin, requested string) (string, AuthContext, error) {
	name, auth, err := ResolveAuthContext(plugin.Manifest, requested)
	if err != nil {
		return "", AuthContext{}, err
	}
	if auth.Type != AuthTypeOAuth2 {
		return "", AuthContext{}, fmt.Errorf("plugin has no selected oauth auth context")
	}
	return name, auth, nil
}

// ValidateOAuthRedirectURI accepts only the frontend plugin OAuth callback and,
// when ORIGIN is configured, requires the same external origin.
func ValidateOAuthRedirectURI(raw string) error {
	redirectURL, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if redirectURL.Scheme != "http" && redirectURL.Scheme != "https" {
		return fmt.Errorf("redirect uri scheme must be http or https")
	}
	if redirectURL.Host == "" {
		return fmt.Errorf("redirect uri must be absolute")
	}
	if redirectURL.Path != "/settings/plugins/oauth/callback" {
		return fmt.Errorf("redirect uri path is not allowed")
	}
	if origin := strings.TrimRight(os.Getenv("ORIGIN"), "/"); origin != "" {
		originURL, err := url.Parse(origin)
		if err != nil {
			return err
		}
		if !strings.EqualFold(redirectURL.Scheme, originURL.Scheme) || !strings.EqualFold(redirectURL.Host, originURL.Host) {
			return fmt.Errorf("redirect uri origin does not match ORIGIN")
		}
	}
	return nil
}

func NewOAuthState(size int) string {
	return randomURLToken(size)
}

func NewOAuthCodeVerifier(size int) string {
	return randomURLToken(size)
}

func PKCEChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// ExchangeOAuthToken performs the host-owned OAuth token exchange or refresh.
// The token endpoint must be allowed by the plugin manifest network policy.
func ExchangeOAuthToken(ctx context.Context, manifest Manifest, authContext AuthContext, auth map[string]any, values map[string]string) (*OAuthTokenResponse, error) {
	operation := OperationOAuthExchange
	if values["grant_type"] != "authorization_code" {
		operation = OperationOAuthRefresh
	}
	tokenURL, err := url.Parse(authContext.TokenURL)
	if err != nil {
		return nil, WithDiagnosticCause(Failure("invalid_configuration", operation), "request_policy_rejected")
	}
	if tokenURL.Scheme != "http" && tokenURL.Scheme != "https" {
		return nil, WithDiagnosticCause(Failure("invalid_configuration", operation), "request_policy_rejected")
	}
	if !OAuthTokenURLAllowed(manifest, tokenURL) {
		return nil, WithDiagnosticCause(Failure("invalid_configuration", operation), "request_policy_rejected")
	}

	clientID := StringFromAny(auth["clientId"])
	clientSecret := StringFromAny(auth[AuthFieldClientSecret])
	if clientID == "" {
		return nil, Failure("invalid_client", operation)
	}

	bodyValues := url.Values{}
	for key, value := range values {
		if value != "" {
			bodyValues.Set(key, value)
		}
	}
	bodyValues.Set("client_id", clientID)
	if authContext.TokenAuth == "" || authContext.TokenAuth == TokenAuthClientSecretPost {
		if clientSecret != "" {
			bodyValues.Set("client_secret", clientSecret)
		}
	}

	var body []byte
	contentType := "application/x-www-form-urlencoded"
	if authContext.TokenRequestFormat == TokenRequestFormatJSON {
		jsonBody := map[string]string{}
		for key, value := range bodyValues {
			if len(value) > 0 {
				jsonBody[key] = value[0]
			}
		}
		var err error
		body, err = json.Marshal(jsonBody)
		if err != nil {
			return nil, err
		}
		contentType = "application/json"
	} else {
		body = []byte(bodyValues.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	if authContext.TokenAuth == TokenAuthClientSecretBasic && clientSecret != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, WithDiagnosticCause(Failure("provider_unavailable", operation), "network_unavailable")
	}
	defer resp.Body.Close()
	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if readErr != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil, WithDiagnosticCause(Failure("provider_unavailable", operation), "network_unavailable")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var rejection struct {
			Error string `json:"error"`
		}
		if readErr == nil && len(respBody) <= 1<<20 {
			_ = json.Unmarshal(respBody, &rejection)
		}
		return nil, PluginFailure{PluginError: ClassifyProviderError(operation, resp.StatusCode,
			resp.Header.Get("Retry-After"), rejection.Error, time.Now())}
	}
	if len(respBody) > 1<<20 {
		return nil, WithDiagnosticCause(PluginFailure{PluginError: PluginError{Code: "invalid_output", Operation: operation, HTTPStatus: resp.StatusCode}}, "token_response_invalid")
	}
	var token OAuthTokenResponse
	token.Raw = append([]byte{}, respBody...)
	if err := json.Unmarshal(respBody, &token); err != nil {
		return nil, WithDiagnosticCause(PluginFailure{PluginError: PluginError{Code: "invalid_output", Operation: operation, HTTPStatus: resp.StatusCode}}, "token_response_invalid")
	}
	if token.AccessToken == "" {
		return nil, WithDiagnosticCause(PluginFailure{PluginError: PluginError{Code: "invalid_output", Operation: operation, HTTPStatus: resp.StatusCode}}, "token_response_invalid")
	}
	// Strava's authorization response identifies its athlete explicitly. Other
	// providers can use providerUserId; absence never invents an account identity.
	if manifest.ID == "strava" && token.ProviderUserID == "" {
		var identity struct {
			Athlete struct {
				ID json.Number `json:"id"`
			} `json:"athlete"`
		}
		if json.Unmarshal(respBody, &identity) == nil {
			if id, parseErr := strconv.ParseInt(string(identity.Athlete.ID), 10, 64); parseErr == nil && id > 0 {
				token.ProviderUserID = strconv.FormatInt(id, 10)
			}
		}
	}
	return &token, nil
}

func OAuthTokenURLAllowed(manifest Manifest, tokenURL *url.URL) bool {
	for _, connector := range manifest.Permissions.Network.Connectors {
		if connector.Type != ConnectorTypePublicAPI {
			continue
		}
		baseURL, basePath, err := NormalizeConnectorBase(connector.FixedBaseURL, "")
		if err != nil {
			continue
		}
		target := ResolvedConnectorTarget{
			Name:                connector.Name,
			Type:                connector.Type,
			BaseURL:             baseURL,
			BasePath:            basePath,
			AllowedPathPrefixes: connector.AllowedPathPrefixes,
		}
		if err := ValidateConnectorURL(target, tokenURL); err == nil {
			return true
		}
	}
	return false
}

// RefreshOAuthToken exchanges outside the transaction and patches tokens only
// under C/T. A concurrent rotation reloads current auth without provider replay.
func RefreshOAuthToken(ctx context.Context, app core.App, plugin LocalPlugin, instance *core.Record, auth map[string]any, contextName string) (map[string]any, error) {
	snapshot := CaptureConnection(instance)
	current, err := CheckConnection(app, snapshot, false)
	if err != nil {
		return auth, err
	}
	tokenRevision := instance.GetString("oauth_token_revision")
	observedError := JSONMapFromRecord(current, "last_error")
	observedBackoff := current.GetString("retry_not_before")
	providerUserID := StringFromAny(auth[AuthFieldProviderUserID])
	if tokenRevision == "" {
		return auth, ErrConnectionConflict
	}
	if current.GetString("oauth_token_revision") != tokenRevision {
		return reloadOAuthAuth(instance, current, plugin.Manifest, contextName)
	}
	if contextName != snapshot.AuthContext {
		return auth, Failure("invalid_configuration", OperationOAuthRefresh)
	}
	if err := CheckOAuthGrantProvenance(plugin.Manifest, auth, contextName); err != nil {
		return auth, err
	}
	_, authContext, err := OAuthContext(plugin, contextName)
	if err != nil {
		return auth, err
	}
	grantType := "refresh_token"
	if authContext.Refresh != nil && authContext.Refresh.GrantType != "" {
		grantType = authContext.Refresh.GrantType
	}
	refreshToken := StringFromAny(auth[AuthFieldRefreshToken])
	if refreshToken == "" {
		return auth, Failure("invalid_grant", OperationOAuthRefresh)
	}
	token, err := ExchangeOAuthToken(ctx, plugin.Manifest, authContext, auth, map[string]string{
		"grant_type":    grantType,
		"refresh_token": refreshToken,
	})
	if err != nil {
		fresh, guardErr := CheckConnection(app, snapshot, false)
		if guardErr != nil {
			return auth, guardErr
		}
		if fresh.GetString("oauth_token_revision") != tokenRevision {
			return reloadOAuthAuth(instance, fresh, plugin.Manifest, contextName)
		}
		return auth, err
	}
	if token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}
	if token.ProviderUserID != "" && providerUserID != "" && token.ProviderUserID != providerUserID {
		return auth, Failure("invalid_grant", OperationOAuthRefresh)
	}
	patch := map[string]any{}
	StoreOAuthToken(patch, contextName, token)
	updated, err := PatchConnection(app, snapshot, false, func(fresh *core.Record) error {
		if fresh.GetString("oauth_token_revision") != tokenRevision {
			return ErrOAuthTokenRevisionConflict
		}
		freshProviderUserID := StringFromAny(JSONMapFromRecord(fresh, "auth")[AuthFieldProviderUserID])
		if providerUserID != "" && freshProviderUserID != providerUserID {
			return ErrConnectionConflict
		}
		if token.ProviderUserID != "" && freshProviderUserID != "" && token.ProviderUserID != freshProviderUserID {
			return Failure("invalid_grant", OperationOAuthRefresh)
		}
		if err := ApplyAuthPatch(fresh, plugin.Manifest, patch); err != nil {
			return err
		}
		fresh.Set("oauth_token_revision", NewConnectionRevision())
		// A refresh only disproves a diagnosis attached to the exact old token
		// revision. Unrelated or historical sync errors remain until full sync.
		diagnosis := JSONMapFromRecord(fresh, "last_error")
		if diagnosis["authContext"] == snapshot.AuthContext && diagnosis["connectionGeneration"] == snapshot.Generation &&
			diagnosis["oauthTokenRevision"] == tokenRevision && diagnosis["operation"] == OperationOAuthRefresh && reflect.DeepEqual(diagnosis, observedError) && fresh.GetString("retry_not_before") == observedBackoff {
			fresh.Set("last_error", map[string]any{})
			fresh.Set("retry_not_before", "")
			if fresh.GetBool("enabled") {
				fresh.Set("status", "configured")
			}
		}
		return nil
	})
	if errors.Is(err, ErrOAuthTokenRevisionConflict) {
		fresh, reloadErr := CheckConnection(app, snapshot, false)
		if reloadErr != nil {
			return auth, reloadErr
		}
		return reloadOAuthAuth(instance, fresh, plugin.Manifest, contextName)
	}
	if err != nil {
		if errors.Is(err, ErrConnectionObsolete) {
			return auth, err
		}
		var failure PluginFailure
		if errors.As(err, &failure) {
			return auth, err
		}
		return auth, WithDiagnosticCause(err, "database_save")
	}
	return reloadOAuthAuth(instance, updated, plugin.Manifest, contextName)
}

var ErrOAuthTokenRevisionConflict = errors.New("oauth token revision changed")

func reloadOAuthAuth(instance, fresh *core.Record, manifest Manifest, selected string) (map[string]any, error) {
	if fresh.GetString("oauth_token_revision") == "" {
		return nil, ErrConnectionConflict
	}
	auth, err := DecryptedOAuthGrant(fresh, manifest, selected)
	if err != nil {
		return nil, err
	}
	instance.Set("oauth_token_revision", fresh.GetString("oauth_token_revision"))
	return auth, nil
}

// StoreOAuthToken normalizes provider token responses into the plugin instance
// auth map used by host injection and future refreshes.
func StoreOAuthToken(auth map[string]any, contextName string, token *OAuthTokenResponse) {
	auth[AuthFieldOAuthContext] = contextName
	auth[AuthFieldAccessToken] = token.AccessToken
	if token.ProviderUserID != "" {
		auth[AuthFieldProviderUserID] = token.ProviderUserID
	}
	if token.RefreshToken != "" {
		auth[AuthFieldRefreshToken] = token.RefreshToken
	}
	if token.TokenType != "" {
		auth[AuthFieldTokenType] = token.TokenType
	}
	if token.Scope != "" {
		auth[AuthFieldScope] = token.Scope
	}
	if token.ExpiresIn > 0 {
		auth[AuthFieldExpiresAt] = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
	}
}

// ClearOAuthToken constructs explicit removals of grant and pending-flow data.
// ApplyAuthPatch consumes the nil instructions on a freshly loaded record.
func ClearOAuthToken(auth map[string]any) {
	for _, key := range slices.Concat(
		[]string{AuthFieldAccessToken, AuthFieldRefreshToken},
		InternalOAuthMetadataFields(),
		InternalOAuthTransientFields(),
	) {
		auth[key] = nil
	}
}

// PluginInputAuth exposes only the active context's guest credentials. The
// accessToken remains available for shipped plugins such as Strava; refresh and
// client secrets plus transient OAuth flow fields stay host-only.
func PluginInputAuth(plugin LocalPlugin, auth map[string]any, selected string) map[string]any {
	out, err := ActiveAuth(plugin.Manifest, auth, selected)
	if err != nil {
		return map[string]any{}
	}
	_, context, _ := EffectiveAuthContext(plugin.Manifest, selected)
	if context.Type == AuthTypeOAuth2 {
		for _, key := range PluginInputAuthBlockedFields() {
			delete(out, key)
		}
	}
	return out
}

// RefreshOAuthAuthIfNeeded resolves the active context before considering a
// refresh, so unused OAuth credentials never trigger a provider call.
func RefreshOAuthAuthIfNeeded(ctx context.Context, app core.App, plugin LocalPlugin, instance *core.Record, auth map[string]any) (map[string]any, error) {
	if instance == nil {
		return auth, ErrConnectionConflict
	}
	name, context, err := EffectiveAuthContext(plugin.Manifest, instance.GetString("auth_context"))
	if err != nil {
		return auth, err
	}
	if context.Type == AuthTypeOAuth2 {
		if instance.GetString("oauth_token_revision") == "" {
			return auth, ErrConnectionConflict
		}
		if err := CheckOAuthGrantProvenance(plugin.Manifest, auth, name); err != nil {
			return auth, err
		}
	}
	if context.Type == AuthTypeOAuth2 && (StringFromAny(auth[AuthFieldAccessToken]) == "" || OAuthNeedsRefresh(auth)) {
		return RefreshOAuthToken(ctx, app, plugin, instance, auth, name)
	}
	return auth, nil
}

func OAuthNeedsRefresh(auth map[string]any) bool {
	expiresAt := StringFromAny(auth[AuthFieldExpiresAt])
	if expiresAt == "" {
		return false
	}
	parsed, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return false
	}
	return time.Until(parsed) < time.Minute
}

func StringFromAny(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func randomURLToken(size int) string {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
