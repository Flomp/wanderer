import { describe, expect, it } from "vitest";
import { plugin_auth_validate, plugin_oauth_revoke, pluginInstances } from "./plugin_instance_store";
import { get } from "svelte/store";

describe("plugin_auth_validate", () => {
    it("passes the exact unsaved auth and connector config to the validation route", async () => {
        const candidate = { pluginId: "configured-session", instanceId: "abcdefghijklmn1", authContext: "session", expected_connection_generation: "revision",
            auth: { email: "replacement@example.org", password: "new-secret" },
            config: { host: { connectors: { service: { baseURL: "https://draft.example", allowPrivate: true } } },
                plugin: { selector: "unsaved" } } };
        let requested = false;
        await plugin_auth_validate(candidate, async (url, options) => {
            requested = true;
            expect(url).toBe("/api/v1/plugin-system/auth/validate");
            expect(options?.method).toBe("POST");
            expect(JSON.parse(String(options?.body))).toEqual(candidate);
            return new Response(JSON.stringify({ ok: true, authContext: "session" }), { status: 200 });
        });
        expect(requested).toBe(true);
        expect(candidate.config.plugin.selector).toBe("unsaved");
    });
});

describe("plugin_oauth_revoke", () => {
    it("returns its committed generation even when the store refresh sees a newer connection", async () => {
        const result = await plugin_oauth_revoke("abcdefghijklmn1", "opening-generation", async (url, options) => {
            if (url === "/api/v1/plugin-system/oauth/revoke") {
                expect(JSON.parse(String(options?.body))).toEqual({ instanceId: "abcdefghijklmn1", expected_connection_generation: "opening-generation" });
                return new Response(JSON.stringify({ ok: true, connection_generation: "own-revoke-generation", updated: "own-revoke-commit" }));
            }
            expect(url).toBe("/api/v1/plugin-instance/abcdefghijklmn1");
            return new Response(JSON.stringify({ id: "abcdefghijklmn1", plugin_id: "oauth", connection_generation: "newer-generation" }));
        });
        expect(result.connection_generation).toBe("own-revoke-generation");
        expect(result.updated).toBe("own-revoke-commit");
        expect(get(pluginInstances).find(item => item.id === "abcdefghijklmn1")?.connection_generation).toBe("newer-generation");
    });
});
