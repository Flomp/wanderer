package pluginsystem

import (
	"os"
	"slices"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"pocketbase/util"
)

// EditableAuthCredentialFields is shared by Settings and draft validation.
// Secret declarations alone specify encryption, rather than editability.
func EditableAuthCredentialFields(manifest Manifest, selected string) map[string]bool {
	fields := map[string]bool{}
	for _, context := range manifest.Auth.Contexts {
		for _, field := range context.Fields {
			fields[field] = true
		}
		if context.Type == AuthTypeAPIKey || context.Type == AuthTypeBearer {
			fields[context.SecretField] = true
		}
		if context.Type == AuthTypeOAuth2 {
			fields[AuthFieldClientSecret] = true
		}
	}
	for _, field := range append(append([]string{AuthFieldRefreshToken}, InternalOAuthMetadataFields()...), InternalOAuthTransientFields()...) {
		delete(fields, field)
	}
	context := manifest.Auth.Contexts[selected]
	if (context.Type != AuthTypeAPIKey && context.Type != AuthTypeBearer) || (context.SecretField != AuthFieldAccessToken && !slices.Contains(context.Fields, AuthFieldAccessToken)) {
		delete(fields, AuthFieldAccessToken)
	}
	delete(fields, "")
	return fields
}

const (
	AuthFieldAccessToken             = "accessToken"
	AuthFieldRefreshToken            = "refreshToken"
	AuthFieldClientSecret            = "clientSecret"
	AuthFieldOAuthState              = "oauthState"
	AuthFieldOAuthCodeVerifier       = "oauthCodeVerifier"
	AuthFieldOAuthRedirectURI        = "oauthRedirectURI"
	AuthFieldOAuthFlowContext        = "oauthFlowContext"
	AuthFieldOAuthActivateOnCallback = "oauthActivateOnCallback"
	AuthFieldProviderUserID          = "providerUserId"
)

func InternalAuthSecretFields() []string {
	return []string{
		AuthFieldAccessToken,
		AuthFieldRefreshToken,
		AuthFieldClientSecret,
		AuthFieldOAuthState,
		AuthFieldOAuthCodeVerifier,
	}
}

func InternalOAuthTransientFields() []string {
	return []string{
		AuthFieldOAuthState,
		AuthFieldOAuthCodeVerifier,
		AuthFieldOAuthRedirectURI,
		AuthFieldOAuthFlowContext,
		AuthFieldOAuthActivateOnCallback,
	}
}

// InternalOAuthMetadataFields are host-managed token metadata, rather than
// editable credentials. Omitted values survive settings saves; explicit nil
// values are consumed as removal instructions by the responsible auth writer.
func InternalOAuthMetadataFields() []string {
	return []string{
		AuthFieldExpiresAt,
		AuthFieldTokenType,
		AuthFieldOAuthContext,
		AuthFieldScope,
		AuthFieldProviderUserID,
	}
}

func PluginInputAuthBlockedFields() []string {
	return []string{
		AuthFieldRefreshToken,
		AuthFieldClientSecret,
		AuthFieldOAuthState,
		AuthFieldOAuthCodeVerifier,
		AuthFieldOAuthRedirectURI,
		AuthFieldOAuthFlowContext,
		AuthFieldOAuthActivateOnCallback,
	}
}

// ApplyAuthPatch merges an internal writer's plaintext patch into the fresh
// record. Omission preserves stored bytes, nil removes a key, and only patch
// secrets are encrypted. Status saves never rewrite credentials.
func ApplyAuthPatch(record *core.Record, manifest Manifest, patch map[string]any) error {
	auth := JSONMapFromRecord(record, "auth")
	secretFields := ManifestAuthSecretFields(manifest)
	for field, value := range patch {
		if value == nil {
			delete(auth, field)
			continue
		}
		if secret, ok := value.(string); ok && secret != "" && secretFields[field] {
			encrypted, err := security.Encrypt([]byte(secret), os.Getenv("POCKETBASE_ENCRYPTION_KEY"))
			if err != nil {
				return err
			}
			value = encrypted
		}
		auth[field] = value
	}
	record.Set("auth", auth)
	return nil
}

// DecryptAuthFields decrypts only operation inputs from the shared encrypted
// field set, preserving inactive credentials and flow secrets as opaque bytes.
func DecryptAuthFields(auth map[string]any, manifest Manifest, fields []string, plaintext map[string]bool) error {
	secretFields := ManifestAuthSecretFields(manifest)
	seen := map[string]bool{}
	for _, field := range fields {
		if seen[field] || plaintext[field] || !secretFields[field] {
			continue
		}
		seen[field] = true
		value, ok := auth[field].(string)
		if !ok || value == "" || !util.LooksLikeEncrypted(value) {
			continue
		}
		secret, err := security.Decrypt(value, os.Getenv("POCKETBASE_ENCRYPTION_KEY"))
		if err != nil {
			if IsLegacyPlaintextAuthSecret(manifest.ID, field, value) {
				continue
			}
			return ErrActiveCredentialsUnreadable
		}
		auth[field] = string(secret)
	}
	return nil
}

func DecryptedOAuthGrant(record *core.Record, manifest Manifest, selected string) (map[string]any, error) {
	auth := JSONMapFromRecord(record, "auth")
	if err := CheckOAuthGrantProvenance(manifest, auth, selected); err != nil {
		return nil, err
	}
	_, context, err := EffectiveAuthContext(manifest, selected)
	if err != nil {
		return nil, err
	}
	fields := append(append([]string{}, context.Fields...), context.SecretFields...)
	fields = append(fields, context.SecretField, AuthFieldAccessToken, AuthFieldRefreshToken, AuthFieldClientSecret)
	if err := DecryptAuthFields(auth, manifest, fields, nil); err != nil {
		return nil, err
	}
	return auth, nil
}

// ManifestAuthSecretFields is the writer's encrypted field set. It includes
// every context, because a stored credential can be shared by active and
// inactive contexts. Readers intersect this set with fields needed by the
// current operation; being an input field alone does not make a value secret.
func ManifestAuthSecretFields(manifest Manifest) map[string]bool {
	fields := map[string]bool{}
	for _, field := range InternalAuthSecretFields() {
		fields[field] = true
	}
	for _, context := range manifest.Auth.Contexts {
		if context.SecretField != "" {
			fields[context.SecretField] = true
		}
		for _, field := range context.SecretFields {
			fields[field] = true
		}
	}
	return fields
}

// IsLegacyPlaintextAuthSecret recognizes the known raw Strava credential
// representation imported by pre-plugin migrations. Forty hex characters also
// form a possible unmarked PocketBase ciphertext (30 decoded bytes), so format
// alone cannot distinguish these values. Callers must prefer a successful
// current-key decryption; only its failure permits this narrow legacy fallback.
// This assumes these three Strava fields use the provider's 40-hex format.
// Other providers, fields and ciphertext formats never receive this exception.
func IsLegacyPlaintextAuthSecret(pluginID, field, value string) bool {
	if pluginID != "strava" || len(value) != 40 {
		return false
	}
	switch field {
	case AuthFieldAccessToken, AuthFieldRefreshToken, AuthFieldClientSecret:
	default:
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}
