import { describe, expect, it } from "vitest";
import { resolvePluginAuthContext, resolveProviderAuth, hasPluginConnectionMaterial, editableAuthFields, pluginAuthDraftInstance } from "./plugin_auth_context";
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
    it("never previews a saved OAuth grant as another method's API key", () => {
        const preview = pluginAuthDraftInstance(plugin, { ...instance, auth: {
            accessToken: "", refreshToken: "", oauthContext: "oauth", clientId: "client",
        } }, "key")!;
        expect(preview.auth).toEqual({ clientId: "client" });
        expect(hasPluginConnectionMaterial(plugin, preview)).toBe(false);
        expect(instance.auth).toEqual({ oauthContext: "oauth", accessToken: "" });
        const legacy = pluginAuthDraftInstance(plugin, { ...instance, auth: { accessToken: "" } }, "key")!;
        expect(legacy.auth).toEqual({});
    });
    it("does not reuse an unattributed token from a removed method after an OAuth-only manifest update", () => {
        const updatedPlugin = { ...plugin, authContexts: { oauth: { type: "oauth2" } } };
        for (const oauthContext of [undefined, ""]) {
            const removed = { ...instance, auth_context: "removed-api-context", auth: { accessToken: "", ...(oauthContext === undefined ? {} : { oauthContext }) } };
            const preview = pluginAuthDraftInstance(updatedPlugin, removed, "oauth")!;
            expect(preview.auth).toEqual({});
            expect(hasPluginConnectionMaterial(updatedPlugin, preview)).toBe(false);
            expect(hasPluginConnectionMaterial(updatedPlugin, pluginAuthDraftInstance(updatedPlugin, { ...removed, auth_context: "" }, "oauth"))).toBe(true);
        }
        const attributed = pluginAuthDraftInstance(updatedPlugin, { ...instance, auth_context: "removed-api-context",
            auth: { accessToken: "", oauthContext: "oauth" } }, "oauth");
        expect(hasPluginConnectionMaterial(updatedPlugin, attributed)).toBe(true);
    });
    it("recognizes OAuth grant evidence after its stored context is removed", () => {
        const cases: Record<string, string>[] = [{ refreshToken: "" }, { oauthContext: "removed-oauth" }, { scope: "" }];
        for (const evidence of cases) {
            const removed = { ...instance, auth_context: "removed-oauth", auth: { accessToken: "", ...evidence } };
            const preview = pluginAuthDraftInstance(plugin, removed, "key")!;
            expect(preview.auth).toEqual({});
            expect(hasPluginConnectionMaterial(plugin, preview)).toBe(false);
        }
    });
    it("preserves a shared API key when switching between non-OAuth methods", () => {
        const apiPlugin = { ...plugin, authContexts: {
            keyA: { type: "api_key", secretField: "accessToken" },
            keyB: { type: "api_key", secretField: "accessToken" },
        } };
        const apiInstance = { ...instance, auth_context: "keyA", auth: { accessToken: "" } };
        const preview = pluginAuthDraftInstance(apiPlugin, apiInstance, "keyB")!;
        expect(preview.auth).toEqual({ accessToken: "" });
        expect(hasPluginConnectionMaterial(apiPlugin, preview)).toBe(true);
    });
    it("keeps encryption declarations separate from the editable input allowlist", () => {
        expect(editableAuthFields({ type: "oauth2", secretFields: ["oauthState", "accessToken"] })).toEqual(["clientSecret"]);
        expect(editableAuthFields({ type: "api_key", secretField: "accessToken" })).toEqual(["accessToken"]);
    });
});
