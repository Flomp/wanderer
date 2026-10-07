package pluginsystem

import (
	"errors"
	"fmt"
)

var ErrActiveCredentialsUnreadable = errors.New("active credentials cannot be decrypted")

// ResolveAuthContext accepts an explicit selection or the sole declared context.
// A missing selection never silently chooses between multiple accounts/modes.
func ResolveAuthContext(manifest Manifest, requested string) (string, AuthContext, error) {
	if requested != "" {
		auth, ok := manifest.Auth.Contexts[requested]
		if !ok {
			return "", AuthContext{}, fmt.Errorf("unknown auth context")
		}
		return requested, auth, nil
	}
	if len(manifest.Auth.Contexts) == 0 {
		return "", AuthContext{}, nil
	}
	if len(manifest.Auth.Contexts) != 1 {
		return "", AuthContext{}, fmt.Errorf("auth context selection is required")
	}
	for name, auth := range manifest.Auth.Contexts {
		return name, auth, nil
	}
	panic("unreachable")
}

func EffectiveAuthContext(manifest Manifest, auth map[string]any) (string, AuthContext, error) {
	selected, exists := auth[AuthFieldOAuthContext]
	if exists && selected != nil {
		name, ok := selected.(string)
		if !ok {
			return "", AuthContext{}, fmt.Errorf("auth context must be a string")
		}
		return ResolveAuthContext(manifest, name)
	}
	return ResolveAuthContext(manifest, "")
}

// ActiveAuth retains only credentials/metadata consumed by the chosen context.
// Pending OAuth flow secrets and credentials for inactive contexts stay host-only.
func ActiveAuth(manifest Manifest, auth map[string]any) (map[string]any, error) {
	name, context, err := EffectiveAuthContext(manifest, auth)
	if err != nil {
		return nil, err
	}
	filtered := AuthForPluginRefresh(auth, context)
	if context.Type == AuthTypeOAuth2 {
		for _, key := range append([]string{AuthFieldAccessToken, AuthFieldRefreshToken, AuthFieldClientSecret, "clientId"}, InternalOAuthMetadataFields()...) {
			if value, ok := auth[key]; ok {
				filtered[key] = value
			}
		}
	}
	if name != "" && context.Type == AuthTypeOAuth2 {
		filtered[AuthFieldOAuthContext] = name
	}
	return filtered, nil
}
