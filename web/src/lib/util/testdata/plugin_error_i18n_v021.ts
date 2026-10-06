// Unchanged translation implementation from wanderer v0.21.0
// (a93566bcd055e5a0a2c2a849cad2f6544fb8f579), kept as a compatibility fixture.
import { get } from "svelte/store";
import { _ } from "svelte-i18n";

const credentialErrorCodes = new Set(["auth_failed"]);
const reconnectErrorCodes = new Set(["invalid_grant", "unauthorized"]);
const unavailableErrorCodes = new Set(["provider_unavailable", "temporary_unavailable"]);
const internalErrorCodes = new Set(["internal_error", "plugin_error"]);

export function translatePluginError(code?: string, message?: string): string {
    const normalizedCode = code?.trim();

    if (normalizedCode && credentialErrorCodes.has(normalizedCode)) {
        return get(_)("wrong-username-or-password");
    }
    if (normalizedCode && reconnectErrorCodes.has(normalizedCode)) {
        return get(_)("plugin-error-reconnect-required");
    }
    if (normalizedCode === "rate_limited") {
        return get(_)("plugin-error-rate-limited");
    }
    if (normalizedCode && unavailableErrorCodes.has(normalizedCode)) {
        return get(_)("plugin-error-provider-unavailable");
    }
    if (normalizedCode === "invalid_request") {
        return get(_)("plugin-error-invalid-request");
    }
    if (normalizedCode && internalErrorCodes.has(normalizedCode)) {
        return get(_)("plugin-error-internal");
    }

    return message?.trim() || get(_)("plugin-setup-error");
}
