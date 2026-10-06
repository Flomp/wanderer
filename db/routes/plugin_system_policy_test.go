package routes

import (
	"bytes"
	"encoding/pem"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pocketbase/pluginsystem"
)

func TestPluginInstancePolicyUsesHostConnectorConfig(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	plugin := pluginsystem.LocalPlugin{Manifest: pluginsystem.Manifest{
		Permissions: pluginsystem.PermissionManifest{
			Network: pluginsystem.NetworkPermissions{
				Connectors: []pluginsystem.ConnectorTargetPermission{{
					Name:              "media",
					Type:              pluginsystem.ConnectorTypeConfigured,
					ConfigKey:         "immich",
					SupportsCustomTLS: true,
				}},
			},
		},
	}}
	config := map[string]any{
		"plugin": map[string]any{
			"after": "2026-01-01",
		},
		"host": map[string]any{
			"connectors": map[string]any{
				"immich": map[string]any{
					"baseURL":      "https://photos.example.test",
					"basePath":     "/immich",
					"allowPrivate": true,
					"tls": map[string]any{
						"mode":     pluginsystem.TLSModeCustomCA,
						"caBundle": string(certificate),
					},
				},
			},
		},
	}

	policy, err := pluginInstancePolicy(plugin, config)
	if err != nil {
		t.Fatal(err)
	}
	connector, ok := policy.Connectors["media"]
	if !ok {
		t.Fatal("expected configured connector to be resolved from host config")
	}
	if connector.BaseURL != "https://photos.example.test" || connector.BasePath != "/immich" {
		t.Fatalf("unexpected connector base: %#v", connector)
	}
	if !connector.AllowPrivate {
		t.Fatal("expected allowPrivate from host connector config")
	}
	if connector.TLS.Mode != pluginsystem.TLSModeCustomCA || string(connector.TLS.CABundle) != string(certificate) {
		t.Fatalf("unexpected TLS config: %#v", connector.TLS)
	}
}

// Invalid local config must be repairable and diagnosable without reflecting
// submitted URLs, credentials, certificates or arbitrary manifest keys.
func TestPluginInstancePolicyRejectsConfigWithSafeDiagnostic(t *testing.T) {
	const secret = "submitted-secret"
	configured := pluginsystem.ConnectorTargetPermission{Name: "media", Type: pluginsystem.ConnectorTypeConfigured, ConfigKey: "media", SupportsStorageRedirects: true}
	required := []pluginsystem.ConfigField{{Key: secret, Type: "string", Required: true}}
	for _, tc := range []struct {
		name, cause string
		connector   pluginsystem.ConnectorTargetPermission
		fields      []pluginsystem.ConfigField
		config      map[string]any
	}{
		{name: "missing connector", cause: "connector_resolution_failed", connector: configured, config: map[string]any{}},
		{name: "invalid shape", cause: "config_shape_invalid", config: map[string]any{"host": secret}},
		{name: "required config", cause: "required_config_missing", fields: required, config: map[string]any{}},
		{name: "wrong type", cause: "config_value_type_invalid", fields: []pluginsystem.ConfigField{{Key: secret, Type: "boolean"}}, config: map[string]any{"plugin": map[string]any{secret: secret}}},
		{name: "invalid option", cause: "config_option_invalid", fields: []pluginsystem.ConfigField{{Key: secret, Type: "string", Options: []pluginsystem.ConfigFieldOption{{Value: "allowed"}}}}, config: map[string]any{"plugin": map[string]any{secret: secret}}},
		{name: "credential URL", cause: "connector_resolution_failed", connector: configured, config: connectorConfigForDiagnostic(map[string]any{"baseURL": "https://user:" + secret + "@photos.test"})},
		{name: "unknown TLS", cause: "tls_mode_unsupported", connector: configured, config: connectorConfigForDiagnostic(map[string]any{"baseURL": "https://photos.test", "tls": map[string]any{"mode": secret}})},
		{name: "disallowed custom TLS", cause: "tls_custom_not_allowed", connector: configured, config: connectorConfigForDiagnostic(map[string]any{"baseURL": "https://photos.test", "tls": map[string]any{"mode": pluginsystem.TLSModeCustomCA, "caBundle": secret}})},
		{name: "bad certificate", cause: "tls_ca_invalid", connector: pluginsystem.ConnectorTargetPermission{Name: "media", Type: pluginsystem.ConnectorTypeConfigured, ConfigKey: "media", SupportsCustomTLS: true}, config: connectorConfigForDiagnostic(map[string]any{"baseURL": "https://photos.test", "tls": map[string]any{"mode": pluginsystem.TLSModeCustomCA, "caBundle": secret}})},
		{name: "bad storage URL", cause: "storage_origin_invalid", connector: configured, config: connectorConfigForDiagnostic(map[string]any{"baseURL": "https://photos.test", "storageOrigins": map[string]any{"storage": map[string]any{"baseURL": secret}}})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manifest := pluginsystem.Manifest{ConfigSchema: tc.fields}
			if tc.connector.Name != "" {
				manifest.Permissions.Network.Connectors = []pluginsystem.ConnectorTargetPermission{tc.connector}
			}
			policy, err := pluginInstancePolicy(pluginsystem.LocalPlugin{Manifest: manifest}, tc.config)
			if err == nil || len(policy.Connectors) != 0 {
				t.Fatalf("invalid configuration produced policy: %#v, %v", policy, err)
			}
			safe := pluginsystem.PublicPluginError(err)
			if safe.Code != "invalid_configuration" || safe.Operation != pluginsystem.OperationHostHTTP {
				t.Fatalf("unexpected public error: %#v", safe)
			}
			var output bytes.Buffer
			pluginsystem.LogPluginFailure(slog.New(slog.NewJSONHandler(&output, nil)), "plugin configuration rejected", err)
			if !strings.Contains(output.String(), `"cause":"`+tc.cause+`"`) || !strings.Contains(output.String(), `"class":"configuration"`) {
				t.Fatalf("missing config diagnostic: %s", output.String())
			}
			if strings.Contains(output.String(), secret) || strings.Contains(output.String(), "https://") || strings.Contains(err.Error(), secret) {
				t.Fatalf("submitted config reached diagnostics: %s / %v", output.String(), err)
			}
		})
	}
}

func connectorConfigForDiagnostic(config map[string]any) map[string]any {
	return map[string]any{"host": map[string]any{"connectors": map[string]any{"media": config}}}
}
