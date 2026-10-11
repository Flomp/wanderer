import type { RequestEvent } from "@sveltejs/kit";
import PocketBase from "pocketbase";
import { describe, expect, it, vi } from "vitest";
import { POST } from "./+server";

const assets = [
    { id: "asset0000000001", collectionId: "pbc_assets", collectionName: "assets", type: "photo", file: "first.jpg", storage_mode: "copy" },
    { id: "asset0000000002", collectionId: "pbc_assets", collectionName: "assets", type: "photo", file: "second.jpg", storage_mode: "copy" },
];
const links = assets.map((asset, i) => ({
    id: `link00000000000${i + 1}`,
    trail: "trail0000000001",
    asset: asset.id,
    is_thumbnail: i === 1,
    expand: { asset },
}));

// Like PocketBase, the update response only carries the requested expands.
function pocketBaseFetch() {
    return vi.fn(async (url: RequestInfo | URL, _options?: RequestInit) => {
        const expand = new URL(String(url)).searchParams.get("expand")?.split(",") ?? [];
        const record = {
            id: "trail0000000001",
            name: "Renamed",
            date: "2026-10-01 00:00:00.000Z",
            thumbnail: 0,
            ...(expand.includes("trail_assets_via_trail.asset") ? { expand: { trail_assets_via_trail: links } } : {}),
        };
        return new Response(JSON.stringify(record), { headers: { "Content-Type": "application/json" } });
    });
}

function trailUpdateEvent(query: string, fetch: typeof globalThis.fetch): RequestEvent {
    const url = new URL(`https://wanderer.example/api/v1/trail/trail0000000001${query}`);
    return {
        url,
        request: new Request(url, { method: "POST", body: JSON.stringify({ name: "Renamed" }) }),
        params: { id: "trail0000000001" },
        locals: { pb: new PocketBase("https://pocketbase.example") },
        fetch,
    } as unknown as RequestEvent;
}

describe("trail update endpoint", () => {
    // The exact expand also keeps waypoint and summit log expands out of the
    // response, which the trail panel would merge over its loaded ones.
    it.each([
        ["", "trail_assets_via_trail.asset"],
        ["?expand=tags", "tags,trail_assets_via_trail.asset"],
    ])("answers with the trail photos for the query '%s'", async (query, expand) => {
        const fetch = pocketBaseFetch();

        const response = await POST(trailUpdateEvent(query, fetch));

        expect(response.status).toBe(200);
        expect(fetch).toHaveBeenCalledOnce();
        expect(new URL(String(fetch.mock.calls[0][0])).searchParams.get("expand")).toBe(expand);
        const trail = await response.json();
        expect(trail.photos).toEqual([
            "/api/v1/files/pbc_assets/asset0000000001/first.jpg",
            "/api/v1/files/pbc_assets/asset0000000002/second.jpg",
        ]);
        expect(trail.thumbnail).toBe(1);
    });
});
