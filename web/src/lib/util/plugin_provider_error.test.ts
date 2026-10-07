import { describe, expect, it } from "vitest";
import { addMessages, init } from "svelte-i18n";
import en from "$lib/i18n/locales/en.json";
import de from "$lib/i18n/locales/de.json";
import { APIError } from "$lib/util/api_util";
import { pluginErrorKey, translatePluginAPIError, translatePluginError } from "./plugin_error_i18n";

addMessages("en", en);
init({ fallbackLocale: "en", initialLocale: "en" });

describe("safe plugin error display", () => {
    it.each([
        ["invalid_client", "plugin-error-client-credentials"],
        ["credentials_unreadable", "plugin-error-credentials-unreadable"],
        ["invalid_configuration", "plugin-error-invalid-configuration"],
        ["not_importable", "plugin-error-not-importable"],
        ["not_found", "plugin-error-not-found"],
        ["unsupported_protocol", "plugin-error-unsupported-protocol"],
        ["unauthorized_client", "plugin-error-client-credentials"],
        ["invalid_scope", "plugin-error-configuration"],
        ["unsupported_grant_type", "plugin-error-configuration"],
        ["access_denied", "plugin-error-access-denied"],
        ["runtime_timeout", "plugin-error-timeout"],
        ["invalid_output", "plugin-error-invalid-output"],
        ["runtime_error", "plugin-error-internal"],
        ["auth_failed", "wrong-username-or-password"],
        ["invalid_grant", "plugin-error-reconnect-required"],
        ["rate_limited", "plugin-error-rate-limited"],
        ["provider_unavailable", "plugin-error-provider-unavailable"],
    ])("localizes %s without showing provider diagnostics", (code, key) => {
        expect(pluginErrorKey(code)).toBe(key);
        expect(Object.hasOwn(en, key)).toBe(true);
        expect(Object.hasOwn(de, key)).toBe(true);
        expect(translatePluginError(code, "provider-body-secret")).toBe(en[key as keyof typeof en]);
    });

    it("reads the real PocketBase operation error and the proxy's nested body", () => {
        const body = { data: { operation: { code: "invalid_scope", message: "provider-body-secret", params: { operation: "oauth_exchange" } } } };
        for (const detail of [body, { detail: body }, JSON.stringify(body)]) {
            expect(translatePluginAPIError(new APIError(400, "Plugin operation failed.", detail), "fallback")).toBe(en["plugin-error-configuration"]);
        }
    });

    it.each([
        ["Failed to read trail GPX.", "error-sending-trail"],
        ["Failed to save plugin connection.", "error-starting-oauth"],
        ["Failed to save plugin connection.", "error-completing-oauth"],
    ])("keeps the action fallback for local host failure: %s", (message, key) => {
        expect(typeof en[key as keyof typeof en]).toBe("string");
        expect(typeof de[key as keyof typeof de]).toBe("string");
        const fallback = en[key as keyof typeof en] as string;
        const body = { status: 400, message, data: {} };
        for (const detail of [body, { detail: body }, JSON.stringify(body)]) {
            expect(translatePluginAPIError(new APIError(400, message, detail), fallback)).toBe(fallback);
        }
    });

    it("never displays raw unknown or historical diagnostics", () => {
        expect(translatePluginError("future-provider-secret", "provider-body-secret")).toBe(en["plugin-error-internal"]);
        expect(translatePluginAPIError(new Error("https://user:secret@provider.test/?token=secret"), "safe fallback")).toBe("safe fallback");
        expect(translatePluginError("toString", "provider-body-secret")).toBe(en["plugin-error-internal"]);
    });
});
