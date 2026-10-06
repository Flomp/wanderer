import { describe, expect, it } from "vitest";
import { resolvePluginAuthContext, resolveProviderAuth } from "./plugin_auth_context";
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

    it("uses the stored context to select the correct fields in a multi-context dialog", () => {
        const plugin: PluginProvider = {
            id: "test", type: "trails", name: "Test", status: "available",
            auth: { type: "unresolved" },
            authContexts: {
                session: { type: "session", fields: ["email", "password"] },
                oauth: { type: "oauth2", fields: ["clientId", "clientSecret"] },
            },
        };
        expect(resolveProviderAuth(plugin, { auth: { oauthContext: "oauth" } })).toEqual({
            name: "oauth", auth: plugin.authContexts!.oauth,
        });
        expect(() => resolveProviderAuth(plugin)).toThrow("ambiguous");
    });
});
