package pluginsystem

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

func ClearPendingOAuthFlow(auth map[string]any) {
	for _, key := range InternalOAuthTransientFields() {
		delete(auth, key)
	}
}

func ClearOAuthGrant(auth map[string]any) {
	for _, key := range append([]string{AuthFieldAccessToken, AuthFieldRefreshToken}, InternalOAuthMetadataFields()...) {
		delete(auth, key)
	}
	ClearPendingOAuthFlow(auth)
}

func ResetConnection(record *core.Record) {
	record.Set("connection_generation", NewConnectionRevision())
	record.Set("state", map[string]any{})
	record.Set("last_error", map[string]any{})
	record.Set("retry_not_before", "")
	var auth map[string]any
	if err := record.UnmarshalJSONField("auth", &auth); err == nil && auth != nil {
		ClearPendingOAuthFlow(auth)
		record.Set("auth", auth)
	}
}

// LocalConnectionStatus establishes local prerequisites only. Reading and
// upgrade migrations neither decrypt credentials nor contact the provider.
func LocalConnectionStatus(manifest Manifest, selected string, auth map[string]any, enabled bool, now time.Time) string {
	if !enabled {
		return "disabled"
	}
	if len(manifest.Auth.Contexts) == 0 {
		return "configured"
	}
	ctx, exists := manifest.Auth.Contexts[selected]
	if selected == "" || !exists || ValidateAuthContext(selected, ctx) != nil || auth == nil {
		return "error"
	}
	fields := append(append([]string{}, ctx.Fields...), ctx.SecretFields...)
	if ctx.SecretField != "" {
		fields = append(fields, ctx.SecretField)
	}
	for _, key := range fields {
		if value, exists := auth[key]; exists {
			if _, ok := value.(string); !ok {
				return "error"
			}
		}
	}
	present := func(key string) bool { value, _ := auth[key].(string); return value != "" }
	switch ctx.Type {
	case AuthTypeSession:
		for _, key := range fields {
			if !present(key) {
				return "needs_auth"
			}
		}
		return "configured"
	case AuthTypeAPIKey, AuthTypeBearer:
		if !present(ctx.SecretField) {
			return "needs_auth"
		}
		return "configured"
	case AuthTypeOAuth2:
		for _, key := range []string{"accessToken", "refreshToken"} {
			if value, exists := auth[key]; exists {
				if _, ok := value.(string); !ok {
					return "error"
				}
			}
		}
		access, refresh := present("accessToken"), present("refreshToken")
		if !access && !refresh {
			return "needs_auth"
		}
		if CheckOAuthGrantProvenance(manifest, auth, selected) != nil {
			return "needs_auth"
		}
		expires, known := localOAuthExpiry(auth["expiresAt"])
		if access && (!known || expires.After(now)) {
			return "configured"
		}
		if ctx.Refresh != nil && ctx.Refresh.Mode == "host" && refresh {
			complete := true
			for _, key := range fields {
				if key != "accessToken" && key != "refreshToken" && !present(key) {
					complete = false
				}
			}
			if complete {
				return "configured"
			}
		}
		return "needs_reauth"
	}
	return "error"
}

func localOAuthExpiry(raw any) (time.Time, bool) {
	var seconds int64
	switch value := raw.(type) {
	case float64:
		seconds = int64(value)
	case int64:
		seconds = value
	case int:
		seconds = int64(value)
	case json.Number:
		seconds, _ = value.Int64()
	case string:
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			return parsed, true
		}
		seconds, _ = strconv.ParseInt(value, 10, 64)
	}
	if seconds <= 0 {
		return time.Time{}, false
	}
	return time.Unix(seconds, 0), true
}
