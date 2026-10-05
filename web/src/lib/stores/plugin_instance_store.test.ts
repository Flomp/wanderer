import { describe, expect, it } from "vitest";
import { plugin_auth_validate } from "./plugin_instance_store";

describe("plugin_auth_validate", () => {
    it("passes the exact unsaved auth and connector config to the validation route", async () => {
        const candidate = { pluginId: "configured-session", instanceId: "abcdefghijklmn1", authContext: "session",
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
