import type { RequestEvent } from "@sveltejs/kit";
import { describe, expect, it, vi } from "vitest";
import { PUT } from "./+server";

function uploadEvent(fileMetadata: unknown[], metadata?: Record<string, unknown>) {
    const created: FormData[] = [];
    const collection = vi.fn((name: string) => ({
        getOne: vi.fn().mockResolvedValue({ id: "trail0000000001", author: "actor0000000001" }),
        getFullList: vi.fn().mockResolvedValue([]),
        create: vi.fn(async (data: FormData) => {
            if (name === "assets") {
                created.push(data);
            }
            return { id: `asset000000000${created.length}` };
        }),
    }));
    const body = new FormData();
    body.append("trail", "trail0000000001");
    for (const [index] of fileMetadata.entries()) {
        body.append("files", new File(["no exif"], `photo${index}.jpg`, { type: "image/jpeg" }));
    }
    body.append("fileMetadata", JSON.stringify(fileMetadata));
    if (metadata) {
        body.append("metadata", JSON.stringify(metadata));
    }
    const event = {
        request: new Request("http://localhost/api/v1/assets", { method: "PUT", body }),
        locals: { user: { actor: "actor0000000001" }, pb: { collection, filter: vi.fn(() => "") } },
    } as unknown as RequestEvent;
    return { event, created };
}

describe("photo upload", () => {
    it("marks taken_at only when it is the time the photo was taken", async () => {
        const { event, created } = uploadEvent(
            [
                { takenAt: "2026-10-03T08:15:00+02:00", takenAtSource: "exif" },
                { takenAt: "2026-10-03T08:15:00Z" },
                { coordinates: { lat: 46, lon: 8 }, takenAtSource: "exif" },
            ],
            { source: "test" },
        );

        expect((await PUT(event)).status).toBe(200);
        expect(created.map((asset) => asset.get("taken_at"))).toEqual([
            "2026-10-03T06:15:00.000Z",
            "2026-10-03T08:15:00.000Z",
            null,
        ]);
        expect(created.map((asset) => JSON.parse(String(asset.get("metadata"))))).toEqual([
            { source: "test", capture_time: { source: "exif", taken_at: "2026-10-03T06:15:00.000Z" } },
            { source: "test" },
            { source: "test" },
        ]);
    });
});
