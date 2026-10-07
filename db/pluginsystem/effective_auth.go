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

func EffectiveAuthContext(manifest Manifest, selected string) (string, AuthContext, error) {
	if selected == "" && len(manifest.Auth.Contexts) > 0 {
		return "", AuthContext{}, Failure("invalid_configuration", "")
	}
	return ResolveAuthContext(manifest, selected)
}

// CheckOAuthGrantProvenance prevents a context selection from borrowing another
// context's grant. Missing historical provenance is compatible only with the
// sole OAuth context; it never supplies the active selection.
func CheckOAuthGrantProvenance(manifest Manifest, auth map[string]any, selected string) error {
	_, context, err := EffectiveAuthContext(manifest, selected)
	if err != nil || context.Type != AuthTypeOAuth2 {
		return Failure("invalid_configuration", OperationOAuthRefresh)
	}
	provenance, exists := auth[AuthFieldOAuthContext]
	if !exists || provenance == "" {
		if len(manifest.Auth.Contexts) == 1 {
			return nil
		}
		return Failure("invalid_configuration", OperationOAuthRefresh)
	}
	name, ok := provenance.(string)
	if !ok || name != selected {
		return Failure("invalid_configuration", OperationOAuthRefresh)
	}
	return nil
}

// ActiveAuth retains only credentials/metadata consumed by the chosen context.
// Pending OAuth flow secrets and credentials for inactive contexts stay host-only.
func ActiveAuth(manifest Manifest, auth map[string]any, selected string) (map[string]any, error) {
	name, context, err := EffectiveAuthContext(manifest, selected)
	if err != nil {
		return nil, err
	}
	filtered := AuthForPluginRefresh(auth, context)
	if context.Type == AuthTypeOAuth2 {
		if err := CheckOAuthGrantProvenance(manifest, auth, selected); err != nil {
			return nil, err
		}
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
