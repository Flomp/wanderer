import { get } from "svelte/store";
import { _ } from "svelte-i18n";
import { APIError } from "$lib/util/api_util";

type PluginErrorLike = {
    code?: unknown;
    message?: unknown;
};

// This allowlist maps provider/runtime categories without ever displaying an
// untrusted guest message, response body or URL (including historical errors).
const pluginErrorKeys = new Map<string, string>([
    ["auth_failed", "wrong-username-or-password"],
    ["credentials_unreadable", "plugin-error-credentials-unreadable"],
    ["invalid_configuration", "plugin-error-invalid-configuration"],
    ["invalid_grant", "plugin-error-reconnect-required"],
    ["unauthorized", "plugin-error-reconnect-required"],
    ["invalid_client", "plugin-error-client-credentials"],
    ["unauthorized_client", "plugin-error-client-credentials"],
    ["invalid_request", "plugin-error-invalid-request"],
    ["invalid_scope", "plugin-error-configuration"],
    ["unsupported_grant_type", "plugin-error-configuration"],
    ["access_denied", "plugin-error-access-denied"],
    ["rate_limited", "plugin-error-rate-limited"],
    ["provider_unavailable", "plugin-error-provider-unavailable"],
    ["temporary_unavailable", "plugin-error-provider-unavailable"],
    ["runtime_error", "plugin-error-internal"],
    ["runtime_timeout", "plugin-error-timeout"],
    ["invalid_output", "plugin-error-invalid-output"],
    ["internal_error", "plugin-error-internal"],
    ["plugin_error", "plugin-error-internal"],
    ["not_importable", "plugin-error-not-importable"],
    ["not_found", "plugin-error-not-found"],
    ["unsupported_protocol", "plugin-error-unsupported-protocol"],
    ["plugin_instance_busy", "plugin-error-busy"],
    ["plugin_connection_conflict", "plugin-error-connection-conflict"],
    ["plugin_oauth_flow_conflict", "plugin-error-oauth-flow-conflict"],
    ["plugin_instance_exists", "plugin-error-instance-exists"],
    ["plugin_instance_disabled", "plugin-error-instance-disabled"],
    ["connection_generation_required", "plugin-error-connection-conflict"],
]);

export function pluginErrorKey(code: unknown): string {
    return typeof code === "string"
        ? pluginErrorKeys.get(code.trim()) ?? "plugin-error-internal"
        : "plugin-error-internal";
}

const pluginSetupErrorKeys = new Map<string, string>([
    ["manifest_missing", "plugin-setup-error-manifest-missing"],
    ["manifest_unreadable", "plugin-setup-error-manifest-unreadable"],
    ["manifest_invalid", "plugin-setup-error-manifest-invalid"],
    ["runtime_entrypoint_invalid", "plugin-setup-error-runtime-entrypoint-invalid"],
    ["runtime_entrypoint_missing", "plugin-setup-error-runtime-entrypoint-missing"],
    ["runtime_entrypoint_unreadable", "plugin-setup-error-runtime-entrypoint-unreadable"],
    ["setup_failed", "plugin-setup-error-details"],
]);

const authErrorHints = [
    "auth_failed",
    "login failed",
    "email and password are required",
];

const connectionOperationCodes = new Set(["plugin_instance_busy", "plugin_connection_conflict",
    "plugin_oauth_flow_conflict", "plugin_instance_exists", "plugin_instance_disabled", "connection_generation_required"]);

export function translatePluginError(code?: string, _message?: string): string {
    return get(_)(pluginErrorKey(code));
}

// Setup error codes come from plugin discovery and are deliberately mapped
// through an allowlist so backend diagnostics never become user-facing text.
export function pluginSetupErrorKey(code: unknown): string {
    if (typeof code !== "string") {
        return "plugin-setup-error-details";
    }
    return pluginSetupErrorKeys.get(code) ?? "plugin-setup-error-details";
}

export function translatePluginAPIError(error: unknown, fallback: string): string {
    if (error instanceof APIError) {
        const operationCode = pluginOperationCode(error);
        if (operationCode) return translatePluginError(operationCode);
        const pluginError = extractPluginError(error.detail);
        if (pluginError?.code) {
            if (connectionOperationCodes.has(String(pluginError.code))) return fallback;
            return translatePluginError(String(pluginError.code), stringValue(pluginError.message));
        }

        const raw = `${error.message} ${stringValue(error.detail)}`.toLowerCase();
        if (authErrorHints.some((hint) => raw.includes(hint))) {
            return get(_)("wrong-username-or-password");
        }
    }

    return fallback;
}

// Only the operation item in the PocketBase envelope controls conflict UX.
// Other nested field-validation codes never become connection conflicts.
export function pluginOperationCode(error: unknown): string | undefined {
    if (!(error instanceof APIError)) return undefined;
    const code = error.detail?.data?.operation?.code;
    return typeof code === "string" ? code : undefined;
}

export function pluginOperationNeedsReload(error: unknown): boolean {
    return ["plugin_connection_conflict", "plugin_oauth_flow_conflict", "plugin_instance_disabled",
        "plugin_instance_exists", "connection_generation_required"].includes(pluginOperationCode(error) ?? "");
}

function extractPluginError(value: unknown): PluginErrorLike | undefined {
    if (!value) {
        return undefined;
    }

    if (typeof value === "string") {
        return extractPluginErrorFromString(value);
    }

    if (typeof value !== "object") {
        return undefined;
    }

    const record = value as Record<string, unknown>;
    if (typeof record.code === "string") {
        return record;
    }

    for (const nested of Object.values(record)) {
        const parsed = extractPluginError(nested);
        if (parsed) {
            return parsed;
        }
    }

    return undefined;
}

function extractPluginErrorFromString(value: string): PluginErrorLike | undefined {
    try {
        const parsed = JSON.parse(value);
        return extractPluginError(parsed);
    } catch {
        const match = value.match(/"code"\s*:\s*"([^"]+)"/);
        return match ? { code: match[1] } : undefined;
    }
}

function stringValue(value: unknown): string | undefined {
    if (typeof value === "string") {
        return value;
    }
    if (value == null) {
        return undefined;
    }
    return JSON.stringify(value);
}
