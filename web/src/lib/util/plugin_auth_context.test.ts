import { describe, expect, it } from "vitest";
import { resolvePluginAuthContext, resolveProviderAuth, hasPluginConnectionMaterial, editableAuthFields } from "./plugin_auth_context";
import type { PluginProvider } from "$lib/models/plugin_provider";

describe("effective plugin auth context", () => {
    it("keeps no-auth and unique historical contexts compatible", () => {
        expect(resolvePluginAuthContext({})).toEqual({ auth: { type: "none" } });
        expect(resolvePluginAuthContext({ session: { type: "session" } })).toEqual({
            name: "session", auth: { type: "session" },
        });
    });

    it("requires an explicit choice rather than depending on context order", () => {
        const session = { type: "session" };
        const oauth = { type: "oauth2" };
        for (const contexts of [{ session, oauth }, { oauth, session }]) {
            expect(() => resolvePluginAuthContext(contexts)).toThrow("ambiguous");
            expect(resolvePluginAuthContext(contexts, "oauth")).toEqual({ name: "oauth", auth: oauth });
        }
    });

    it("never replaces an unknown explicit context with the only remaining context", () => {
        expect(() => resolvePluginAuthContext({ session: { type: "session" } }, "removed"))
            .toThrow("Unknown");
        expect(() => resolvePluginAuthContext({}, "removed")).toThrow("Unknown");
    });

    it("does not infer the context of an existing instance from grant provenance or a single manifest context", () => {
        const plugin: PluginProvider = { id: "test", type: "trails", name: "Test", status: "available",
            auth: { type: "session" }, authContexts: { session: { type: "session" } } };
        expect(() => resolveProviderAuth(plugin, { auth: { oauthContext: "session" } })).toThrow("missing");
    });

    it("uses the stored context to select the correct fields in a multi-context dialog", () => {
        const plugin: PluginProvider = {
            id: "test", type: "trails", name: "Test", status: "available",
            auth: { type: "unresolved" },
            authContexts: {
                session: { type: "session", fields: ["email", "password"] },
                oauth: { type: "oauth2", fields: ["clientId", "clientSecret"] },
            },
        };
        expect(resolveProviderAuth(plugin, { auth_context: "oauth", auth: { oauthContext: "session" } })).toEqual({
            name: "oauth", auth: plugin.authContexts!.oauth,
        });
        expect(() => resolveProviderAuth(plugin)).toThrow("ambiguous");
    });
});

describe("redacted connection material", () => {
    const plugin: PluginProvider = { id: "test", type: "trails", name: "Test", status: "available", auth: { type: "oauth2" }, authContexts: {
        oauth: { type: "oauth2", fields: ["clientId", "clientSecret"] },
        key: { type: "api_key", secretField: "accessToken" },
    } };
    const instance = { plugin_id: "test", user: "owner", enabled: false, status: "disabled",
        auth_context: "oauth", auth: { oauthContext: "oauth", accessToken: "" } } as const;
    it("uses placeholder keys independently of enabled, busy and transient status", () => {
        for (const status of ["disabled", "rate_limited", "unavailable"] as const) {
            expect(hasPluginConnectionMaterial(plugin, { ...instance, status, busy: true })).toBe(true);
        }
        expect(hasPluginConnectionMaterial(plugin, { ...instance, auth: { clientSecret: "", clientId: "client" } })).toBe(false);
    });
    it("does not mistake another context's accessToken for an OAuth grant", () => {
        expect(hasPluginConnectionMaterial(plugin, { ...instance, auth: { oauthContext: "key", accessToken: "" } })).toBe(false);
        expect(hasPluginConnectionMaterial(plugin, { ...instance, auth: { accessToken: "" } })).toBe(false);
        expect(hasPluginConnectionMaterial(plugin, { ...instance, auth_context: "key", auth: { accessToken: "" } })).toBe(true);
    });
    it("keeps encryption declarations separate from the editable input allowlist", () => {
        expect(editableAuthFields({ type: "oauth2", secretFields: ["oauthState", "accessToken"] })).toEqual(["clientSecret"]);
        expect(editableAuthFields({ type: "api_key", secretField: "accessToken" })).toEqual(["accessToken"]);
    });
});
