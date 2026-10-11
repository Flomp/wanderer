import type { RequestEvent } from "@sveltejs/kit";
import { describe, expect, it, vi } from "vitest";
import { GET } from "./+server";

function requestEvent(search: string) {
    const url = `http://localhost/api/v1/plugins/assets/immich/thumbnail/asset-1${search}`;
    return {
        params: { plugin: "immich", id: "asset-1" },
        url: new URL(url),
        request: new Request(url),
        locals: { pb: { authStore: { token: "" }, buildURL: (path: string) => `http://db/${path}` } },
        fetch: vi.fn().mockResolvedValue(new Response("thumbnail")),
    } as unknown as RequestEvent;
}

describe("asset plugin thumbnail proxy", () => {
    it.each([
        ["?thumb=300x300", "http://db/plugins/assets/immich/thumbnail/asset-1?thumb=300x300"],
        ["", "http://db/plugins/assets/immich/thumbnail/asset-1"],
    ])("proxies %j to %s", async (search, backendURL) => {
        const event = requestEvent(search);
        expect((await GET(event)).status).toBe(200);
        expect(event.fetch).toHaveBeenCalledWith(backendURL, { headers: {} });
    });

    it("rejects a size the backend cannot read", async () => {
        const event = requestEvent("?thumb=small");
        await expect(GET(event)).rejects.toMatchObject({ status: 400 });
        expect(event.fetch).not.toHaveBeenCalled();
    });
});
