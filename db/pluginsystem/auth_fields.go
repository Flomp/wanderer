package pluginsystem

const (
	AuthFieldAccessToken       = "accessToken"
	AuthFieldRefreshToken      = "refreshToken"
	AuthFieldClientSecret      = "clientSecret"
	AuthFieldOAuthState        = "oauthState"
	AuthFieldOAuthCodeVerifier = "oauthCodeVerifier"
	AuthFieldOAuthRedirectURI  = "oauthRedirectURI"
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
	}
}

// InternalOAuthMetadataFields are host-managed token metadata, rather than
// editable credentials. Omitted values survive settings saves; explicit nil
// values are consumed as removal instructions by the instance update hook.
func InternalOAuthMetadataFields() []string {
	return []string{
		AuthFieldExpiresAt,
		AuthFieldTokenType,
		AuthFieldOAuthContext,
		AuthFieldScope,
	}
}

func PluginInputAuthBlockedFields() []string {
	return []string{
		AuthFieldRefreshToken,
		AuthFieldClientSecret,
		AuthFieldOAuthState,
		AuthFieldOAuthCodeVerifier,
		AuthFieldOAuthRedirectURI,
	}
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
