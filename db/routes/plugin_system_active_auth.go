package routes

import (
	"os"
	"slices"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"pocketbase/pluginsystem"
	"pocketbase/util"
)

// Preserve inactive stored fields for the legacy refresh writer, but decrypt
// only active fields encrypted by the manifest writer. Runtime/guest
// projections filter inactive fields out.
func decryptedActiveInstanceAuth(plugin pluginsystem.LocalPlugin, instance *core.Record) (map[string]any, error) {
	auth := pluginsystem.JSONMapFromRecord(instance, "auth")
	_, context, err := pluginsystem.EffectiveAuthContext(plugin.Manifest, auth)
	if err != nil {
		return nil, err
	}
	if err := decryptActiveAuth(auth, plugin.Manifest, context, nil); err != nil {
		return nil, err
	}
	return auth, nil
}

func decryptActiveAuth(auth map[string]any, manifest pluginsystem.Manifest, context pluginsystem.AuthContext, plaintext map[string]bool) error {
	fields := slices.Concat(context.Fields, context.SecretFields)
	if context.SecretField != "" {
		fields = append(fields, context.SecretField)
	}
	if context.Type == pluginsystem.AuthTypeOAuth2 {
		fields = append(fields, pluginsystem.AuthFieldAccessToken, pluginsystem.AuthFieldRefreshToken, pluginsystem.AuthFieldClientSecret)
	}
	return decryptOperationAuthFields(auth, manifest, fields, plaintext)
}

func decryptOperationAuthFields(auth map[string]any, manifest pluginsystem.Manifest, fields []string, plaintext map[string]bool) error {
	secretFields := pluginsystem.ManifestAuthSecretFields(manifest)
	seen := map[string]bool{}
	for _, key := range fields {
		if seen[key] {
			continue
		}
		seen[key] = true
		if plaintext[key] || !secretFields[key] {
			continue
		}
		value, ok := auth[key].(string)
		if !ok || value == "" || !util.LooksLikeEncrypted(value) {
			continue
		}
		secret, err := security.Decrypt(value, os.Getenv("POCKETBASE_ENCRYPTION_KEY"))
		if err != nil {
			if pluginsystem.IsLegacyPlaintextAuthSecret(manifest.ID, key, value) {
				continue
			}
			return pluginsystem.ErrActiveCredentialsUnreadable
		}
		auth[key] = string(secret)
	}
	return nil
}

// OAuth start needs only client credentials; the callback additionally needs
// its pending state/PKCE. Existing grant and inactive-context secrets are not
// prerequisites of either operation and stay in their stored representation.
func decryptedOAuthFlowAuth(instance *core.Record, manifest pluginsystem.Manifest, context pluginsystem.AuthContext, callback bool) (map[string]any, error) {
	auth := pluginsystem.JSONMapFromRecord(instance, "auth")
	fields := []string{}
	for _, field := range append(slices.Concat(context.Fields, context.SecretFields), context.SecretField) {
		if field != "" && field != pluginsystem.AuthFieldAccessToken && field != pluginsystem.AuthFieldRefreshToken {
			fields = append(fields, field)
		}
	}
	fields = append(fields, pluginsystem.AuthFieldClientSecret)
	if callback {
		fields = append(fields, pluginsystem.AuthFieldOAuthState)
		if context.PKCE {
			fields = append(fields, pluginsystem.AuthFieldOAuthCodeVerifier)
		}
	}
	if err := decryptOperationAuthFields(auth, manifest, fields, nil); err != nil {
		return nil, err
	}
	return auth, nil
}
