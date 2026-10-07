import { describe, expect, it, vi } from "vitest";
import { ClientResponseError } from "pocketbase";
import type { RequestEvent } from "@sveltejs/kit";
import { PUT } from "./+server";
import { POST } from "./[id]/+server";

function event(body: unknown, send = vi.fn().mockResolvedValue({ id: "abcdefghijklmn1" })) {
    return { request: new Request("http://wanderer.test/api/v1/plugin-instance", {
        method: "POST", body: JSON.stringify(body), headers: { "content-type": "application/json" },
    }), params: { id: "abcdefghijklmn1" }, locals: { pb: { send } }, fetch } as unknown as RequestEvent;
}

describe("host plugin instance proxies", () => {
    it("creates through the host without a client-supplied owner", async () => {
        const body = { plugin_id: "session", authContext: "account", auth: { password: "replacement" } };
        const e = event(body);
        expect((await PUT(e)).status).toBe(200);
        expect(e.locals.pb.send).toHaveBeenCalledExactlyOnceWith("/plugins/instances", { method: "POST", body, fetch });
    });
    it("forwards explicit null and generation guards to the settings writer", async () => {
        const body = { auth: { password: null }, expected_connection_generation: "revision" };
        const e = event(body);
        expect((await POST(e)).status).toBe(200);
        expect(e.locals.pb.send).toHaveBeenCalledExactlyOnceWith("/plugins/instances/abcdefghijklmn1/settings", { method: "PATCH", body, fetch });
    });
    it.each(["user", "status", "state"])("rejects protected %s without a host write", async key => {
        const e = event({ [key]: "value" });
        expect((await POST(e)).status).toBe(400);
        expect(e.locals.pb.send).not.toHaveBeenCalled();
    });
    it.each(["plugin_instance_busy", "plugin_connection_conflict", "plugin_instance_exists", "plugin_instance_disabled", "plugin_oauth_flow_conflict"])
        ("preserves the exact safe409 envelope for %s", async code => {
            const body = { status: 409, message: "Safe public message", data: { operation: { code, message: "Safe operation message" } } };
            const send = vi.fn().mockRejectedValue(new ClientResponseError({ status: 409, response: body,
                originalError: new Error("private-transport-sentinel") }));
            const response = await POST(event({}, send));
            expect(response.status).toBe(409);
            expect(await response.json()).toEqual(body);
        });
});
