package routes

import (
	"crypto/x509"
	"encoding/base64"
	"strings"

	"pocketbase/pluginsystem"
)

// pluginConfigFailure exposes a repairable code and a host-owned diagnostic
// identifier. User config, URLs and certificate data never become error text.
func pluginConfigFailure(cause string) error {
	return pluginsystem.WithDiagnosticCause(pluginsystem.Failure("invalid_configuration", pluginsystem.OperationHostHTTP), cause)
}

// pluginInstancePolicy validates effective config before returning a complete
// policy; callers never receive a partially resolved connector map.
func pluginInstancePolicy(plugin pluginsystem.LocalPlugin, config map[string]any) (pluginsystem.RequestPolicyContext, error) {
	if err := validatePluginConfig(plugin.Manifest, config); err != nil {
		return pluginsystem.RequestPolicyContext{}, pluginsystem.WithDiagnosticCause(err, "config_rejected")
	}
	connectors := map[string]pluginsystem.ResolvedConnectorTarget{}
	hostConfig := pluginHostConfig(config)
	hostConnectors := configMap(configMap(hostConfig, "connectors"), "")

	for _, manifestConnector := range plugin.Manifest.Permissions.Network.Connectors {
		target, err := resolveConnectorTarget(manifestConnector, hostConnectors)
		if err != nil {
			return pluginsystem.RequestPolicyContext{}, err
		}
		connectors[manifestConnector.Name] = target
	}

	return pluginsystem.RequestPolicyContext{Connectors: connectors}, nil
}

func resolveConnectorTarget(manifest pluginsystem.ConnectorTargetPermission, hostConnectors map[string]any) (pluginsystem.ResolvedConnectorTarget, error) {
	target := pluginsystem.ResolvedConnectorTarget{
		Name:                     manifest.Name,
		Type:                     manifest.Type,
		AllowedPathPrefixes:      manifest.AllowedPathPrefixes,
		Auth:                     manifest.Auth,
		SupportsMediaAuth:        manifest.SupportsMediaAuth,
		SupportsStorageRedirects: manifest.SupportsStorageRedirects,
		SupportsCustomTLS:        manifest.SupportsCustomTLS,
		TLS:                      pluginsystem.ConnectorTLSConfig{Mode: pluginsystem.TLSModeSystem},
		StorageOrigins:           map[string]pluginsystem.ResolvedConnectorOrigin{},
	}

	switch manifest.Type {
	case pluginsystem.ConnectorTypePublicAPI:
		baseURL, basePath, err := pluginsystem.NormalizeConnectorBase(manifest.FixedBaseURL, "")
		if err != nil {
			return target, pluginConfigFailure("connector_resolution_failed")
		}
		target.BaseURL = baseURL
		target.BasePath = basePath
		target.AllowPrivate = false
	case pluginsystem.ConnectorTypeConfigured:
		rawConfig := configMap(hostConnectors, manifest.ConfigKey)
		if len(rawConfig) == 0 {
			return target, pluginConfigFailure("connector_resolution_failed")
		}
		baseURL := stringConfig(rawConfig, "baseURL")
		basePath := stringConfig(rawConfig, "basePath")
		normalizedBaseURL, normalizedBasePath, err := pluginsystem.NormalizeConnectorBase(baseURL, basePath)
		if err != nil {
			return target, pluginConfigFailure("connector_resolution_failed")
		}
		target.BaseURL = normalizedBaseURL
		target.BasePath = normalizedBasePath
		target.AllowPrivate = boolConfig(rawConfig, "allowPrivate")
		target.TLS, err = tlsConfig(rawConfig, manifest.SupportsCustomTLS)
		if err != nil {
			return target, err
		}
		if manifest.SupportsStorageRedirects {
			target.StorageOrigins, err = storageOrigins(rawConfig, manifest.SupportsCustomTLS)
			if err != nil {
				return target, err
			}
		}
	default:
		return target, pluginConfigFailure("connector_resolution_failed")
	}
	return target, nil
}

func storageOrigins(rawConfig map[string]any, customAllowed bool) (map[string]pluginsystem.ResolvedConnectorOrigin, error) {
	rawOrigins, err := optionalConfigObject(rawConfig, "storageOrigins")
	if err != nil {
		return nil, err
	}
	origins := map[string]pluginsystem.ResolvedConnectorOrigin{}
	for name, raw := range rawOrigins {
		origin, ok := raw.(map[string]any)
		if !ok {
			return nil, pluginConfigFailure("storage_origin_invalid")
		}
		if err := validateConnectorValues(origin); err != nil {
			return nil, err
		}
		baseURL, basePath, err := pluginsystem.NormalizeConnectorBase(stringConfig(origin, "baseURL"), stringConfig(origin, "basePath"))
		if err != nil {
			return nil, pluginConfigFailure("storage_origin_invalid")
		}
		tls, err := tlsConfig(origin, customAllowed)
		if err != nil {
			return nil, err
		}
		origins[name] = pluginsystem.ResolvedConnectorOrigin{Name: name, BaseURL: baseURL, BasePath: basePath, AllowPrivate: boolConfig(origin, "allowPrivate"), TLS: tls}
	}
	return origins, nil
}

func tlsConfig(raw map[string]any, customAllowed bool) (pluginsystem.ConnectorTLSConfig, error) {
	rawTLS, err := optionalConfigObject(raw, "tls")
	if err != nil {
		return pluginsystem.ConnectorTLSConfig{}, err
	}
	if value, exists := rawTLS["mode"]; exists {
		if _, ok := value.(string); !ok {
			return pluginsystem.ConnectorTLSConfig{}, pluginConfigFailure("tls_mode_unsupported")
		}
	}
	mode := stringConfig(rawTLS, "mode")
	if mode == "" {
		mode = pluginsystem.TLSModeSystem
	}
	if mode != pluginsystem.TLSModeSystem && mode != pluginsystem.TLSModeCustomCA {
		return pluginsystem.ConnectorTLSConfig{}, pluginConfigFailure("tls_mode_unsupported")
	}
	if mode != pluginsystem.TLSModeSystem && !customAllowed {
		return pluginsystem.ConnectorTLSConfig{}, pluginConfigFailure("tls_custom_not_allowed")
	}
	cfg := pluginsystem.ConnectorTLSConfig{Mode: mode}
	if mode == pluginsystem.TLSModeCustomCA {
		ca, ok := rawTLS["caBundle"].(string)
		if !ok {
			return cfg, pluginConfigFailure("tls_ca_invalid")
		}
		if decoded, err := base64.StdEncoding.DecodeString(ca); err == nil {
			cfg.CABundle = decoded
		} else {
			cfg.CABundle = []byte(ca)
		}
		if !x509.NewCertPool().AppendCertsFromPEM(cfg.CABundle) {
			return cfg, pluginConfigFailure("tls_ca_invalid")
		}
	}
	return cfg, nil
}

func optionalConfigObject(raw map[string]any, key string) (map[string]any, error) {
	value, exists := raw[key]
	if !exists {
		return map[string]any{}, nil
	}
	result, ok := value.(map[string]any)
	if !ok {
		return nil, pluginConfigFailure("config_shape_invalid")
	}
	return result, nil
}

func validateConnectorValues(raw map[string]any) error {
	for _, key := range []string{"baseURL", "basePath"} {
		if value, exists := raw[key]; exists {
			if _, ok := value.(string); !ok {
				return pluginConfigFailure("config_value_type_invalid")
			}
		}
	}
	if value, exists := raw["allowPrivate"]; exists {
		if _, ok := value.(bool); !ok {
			return pluginConfigFailure("config_value_type_invalid")
		}
	}
	return nil
}

func validatePluginConfig(manifest pluginsystem.Manifest, config map[string]any) error {
	pluginConfig, err := optionalConfigObject(config, "plugin")
	if err != nil {
		return err
	}
	hostConfig, err := optionalConfigObject(config, "host")
	if err != nil {
		return err
	}
	connectors, err := optionalConfigObject(hostConfig, "connectors")
	if err != nil {
		return err
	}
	for _, connector := range manifest.Permissions.Network.Connectors {
		if connector.Type != pluginsystem.ConnectorTypeConfigured {
			continue
		}
		raw, err := optionalConfigObject(connectors, connector.ConfigKey)
		if err != nil {
			return err
		}
		if err := validateConnectorValues(raw); err != nil {
			return err
		}
	}
	for _, field := range manifest.ConfigSchema {
		value, present := pluginConfig[field.Key]
		if !present || value == nil || value == "" {
			if field.Required {
				return pluginConfigFailure("required_config_missing")
			}
			continue
		}
		switch field.Type {
		case "boolean":
			if _, ok := value.(bool); !ok {
				return pluginConfigFailure("config_value_type_invalid")
			}
		case "number":
			if _, ok := value.(float64); !ok {
				return pluginConfigFailure("config_value_type_invalid")
			}
		default:
			if _, ok := value.(string); !ok {
				return pluginConfigFailure("config_value_type_invalid")
			}
		}
		if len(field.Options) > 0 {
			valid := false
			for _, option := range field.Options {
				if value == option.Value {
					valid = true
					break
				}
			}
			if !valid {
				return pluginConfigFailure("config_option_invalid")
			}
		}
	}
	return nil
}

func configMap(raw map[string]any, key string) map[string]any {
	if key == "" {
		return raw
	}
	value, ok := raw[key]
	if !ok {
		return map[string]any{}
	}
	switch typed := value.(type) {
	case map[string]any:
		return typed
	default:
		return map[string]any{}
	}
}

func stringConfig(raw map[string]any, key string) string {
	value, _ := raw[key].(string)
	return strings.TrimSpace(value)
}

func boolConfig(raw map[string]any, key string) bool {
	value, _ := raw[key].(bool)
	return value
}
