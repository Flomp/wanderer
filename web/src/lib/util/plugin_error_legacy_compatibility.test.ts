import { describe, expect, it } from "vitest";
import { addMessages, init } from "svelte-i18n";
import en from "$lib/i18n/locales/en.json";
import sdkErrors from "../../../../plugins/sdk/testdata/public_errors.json";
import { translatePluginError as legacyTranslate } from "./testdata/plugin_error_i18n_v021";
import { pluginErrorKey, translatePluginError } from "./plugin_error_i18n";

addMessages("en", en);
init({ fallbackLocale: "en", initialLocale: "en" });

describe("new bundle errors on the released v0.21.0 UI", () => {
    // SDK tests assert these fixture messages against serialized SDK errors.
    it.each(sdkErrors.filter(({ code }) => [
        "credentials_unreadable", "invalid_configuration", "invalid_client",
        "unauthorized_client", "invalid_scope", "unsupported_grant_type",
        "access_denied", "runtime_error", "runtime_timeout", "invalid_output",
        "not_importable", "not_found", "unsupported_protocol",
    ].includes(code)))("shows a safe readable fallback for $code", ({ code, message }) => {
        expect(legacyTranslate(code, message)).toBe(message);
        expect(message).not.toBe(code);
        expect(message.length).toBeLessThanOrEqual(200);
        expect(translatePluginError(code, message)).toBe(en[pluginErrorKey(code) as keyof typeof en]);
    });
});
