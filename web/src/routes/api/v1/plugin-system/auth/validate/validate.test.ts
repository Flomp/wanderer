import { describe, expect, it, vi } from "vitest";
import type { RequestEvent } from "@sveltejs/kit";
import { POST } from "./+server";

describe("session validate proxy", () => {
    it("passes the checked draft to the host without dropping configured connector policy", async () => {
        const draft = { pluginId: "session", instanceId: "abcdefghijklmn1", authContext: "account",
            auth: { email: "replacement@example.org", password: "new-secret" },
            config: { host: { connectors: { service: { baseURL: "http://private.example", allowPrivate: true,
                tls: { mode: "customCA", caBundle: "draft-ca" } } } }, plugin: { selector: "unsaved" } } };
        const send = vi.fn().mockResolvedValue({ ok: true, authContext: "account" });
        const event = { request: new Request("http://wanderer.test/api/v1/plugin-system/auth/validate", {
            method: "POST", body: JSON.stringify(draft), headers: { "content-type": "application/json" },
        }), locals: { pb: { send } }, fetch } as unknown as RequestEvent;
        const response = await POST(event);
        expect(response.status).toBe(200);
        expect(send).toHaveBeenCalledExactlyOnceWith("/plugins/auth/validate", { method: "POST", body: draft, fetch });
        expect(await response.json()).toEqual({ ok: true, authContext: "account" });
    });

    it("rejects malformed config before any host request", async () => {
        const send = vi.fn();
        const event = { request: new Request("http://wanderer.test/api/v1/plugin-system/auth/validate", {
            method: "POST", body: JSON.stringify({ pluginId: "session", auth: {}, config: [] }),
        }), locals: { pb: { send } }, fetch } as unknown as RequestEvent;
        const response = await POST(event);
        expect(response.status).toBe(400);
        expect(send).not.toHaveBeenCalled();
    });
});
