import { error, isHttpError, type RequestEvent } from "@sveltejs/kit";
import { getRequest } from "@sveltejs/kit/node";
import { IncomingMessage } from "node:http";
import { Socket } from "node:net";
import PocketBase, { ClientResponseError } from "pocketbase";
import { describe, expect, it, vi } from "vitest";
import { z } from "zod";
import { assertFileField, Collection, handleError, update, upload, uploadCreate, uploadUpdate } from "./api_util";

function uploadEvent(data: FormData) {
    const update = vi.fn().mockResolvedValue({ id: "record000000001" });
    const create = vi.fn().mockResolvedValue({ id: "record000000001" });
    const event = {
        params: { id: "record000000001" },
        url: new URL("https://example.com/api/v1/test/record000000001"),
        request: { formData: async () => data },
        locals: { pb: { collection: vi.fn(() => ({ update, create })) } },
    } as unknown as RequestEvent;
    return { event, update, create };
}

describe("handleError", () => {
    it("preserves SvelteKit HTTP errors", () => {
        let httpError: unknown;
        try {
            error(404, { message: "Actor not found" });
        } catch (caught) {
            httpError = caught;
        }

        expect(isHttpError(httpError, 404)).toBe(true);
        expect(() => handleError(httpError)).toThrow(httpError);
    });

    it.each(["content-length", "chunked"])("preserves body size errors for %s uploads", async (framing) => {
        const incoming = new IncomingMessage(new Socket());
        incoming.method = "POST";
        incoming.url = "/api/v1/user/user00000000001/file";
        incoming.httpVersionMajor = 1;
        incoming.headers = {
            "content-type": "multipart/form-data; boundary=avatar",
            ...(framing === "content-length"
                ? { "content-length": String(2 * 1024 * 1024) }
                : { "transfer-encoding": "chunked" }),
        };

        try {
            const request = await getRequest({
                request: incoming,
                base: "http://localhost:3000",
                bodySizeLimit: 512 * 1024,
            });
            if (framing === "chunked") {
                incoming.push(Buffer.alloc(2 * 1024 * 1024));
                incoming.push(null);
            }
            const uploadError = await request.formData().catch((caught) => caught);

            expect(uploadError).toMatchObject({ status: 413 });
            expect(isHttpError(uploadError)).toBe(false);
            const response = handleError(uploadError);
            expect(response.status).toBe(413);
            expect(await response.json()).toEqual({ message: uploadError.message });
        } finally {
            incoming.destroy();
        }
    });

    it("keeps unexpected errors as server errors", async () => {
        const response = handleError(new Error("Unexpected failure"));

        expect(response.status).toBe(500);
        expect(await response.json()).toEqual({ message: "Unexpected failure" });
    });
});

describe.each(["JSON", "multipart"] as const)("trail %s update cancellation", (format) => {
    it("passes the request signal through the real PocketBase SDK", async () => {
        const controller = new AbortController();
        const fetch = vi.fn(async (_url: RequestInfo | URL, _options?: RequestInit) =>
            new Response(JSON.stringify({ id: "trail0000000001", public: true }), {
                headers: { "Content-Type": "application/json" },
            }),
        );
        const event = trailUpdateEvent(format, controller.signal, fetch);

        await submitTrailUpdate(event, format);

        expect(fetch).toHaveBeenCalledOnce();
        const [url, options] = fetch.mock.calls[0];
        expect(String(url)).toContain("/api/collections/trails/records/trail0000000001");
        expect(String(url)).toContain("expand=category");
        expect(options?.method).toBe("PATCH");
        // Without requestKey:null the SDK replaces this signal with its own.
        expect(options?.signal).toBe(event.request.signal);
    });

    it("aborts the outgoing PocketBase update when the BFF request is canceled", async () => {
        const controller = new AbortController();
        const fetch = vi.fn((_url: RequestInfo | URL, options?: RequestInit) =>
            new Promise<Response>((_resolve, reject) => {
                const abort = () => reject(new DOMException("Aborted", "AbortError"));
                if (options?.signal?.aborted) {
                    abort();
                } else {
                    options?.signal?.addEventListener("abort", abort, { once: true });
                }
            }),
        );
        const event = trailUpdateEvent(format, controller.signal, fetch);
        const result = expect(submitTrailUpdate(event, format)).rejects.toMatchObject({ isAbort: true });
        await vi.waitFor(() => expect(fetch).toHaveBeenCalledOnce());

        controller.abort();

        await result;
        expect(fetch.mock.calls[0][1]?.signal?.aborted).toBe(true);
    });
});

function trailUpdateEvent(format: "JSON" | "multipart", signal: AbortSignal, fetch: typeof globalThis.fetch): RequestEvent {
    const url = new URL("https://wanderer.example/api/v1/trail/trail0000000001?expand=category");
    const formData = new FormData();
    formData.set("public", "true");
    const request = new Request(url, {
        method: "POST",
        signal,
        ...(format === "multipart"
            ? { body: formData }
            : { body: JSON.stringify({ public: true }), headers: { "Content-Type": "application/json" } }),
    });
    return {
        url,
        request,
        params: { id: "trail0000000001" },
        locals: { pb: new PocketBase("https://pocketbase.example") },
        fetch,
    } as unknown as RequestEvent;
}

function submitTrailUpdate(event: RequestEvent, format: "JSON" | "multipart") {
    return format === "multipart"
        ? uploadUpdate(event, Collection.trails)
        : update(event, z.object({ public: z.boolean() }), Collection.trails);
}

describe("assertFileField", () => {
    const fileFields = ["gpx", "photos"] as const;

    it.each(["gpx", "photos", "photos+", "photos-"])("accepts a `%s` part", (name) => {
        const data = new FormData();
        data.append(name, new Blob(["x"]), "x.bin");

        expect(() => assertFileField(data, fileFields)).not.toThrow();
    });

    it("rejects a body with none of the file fields", () => {
        const data = new FormData();
        data.append("file", new Blob(["x"]), "track.gpx");
        data.append("name", "Renamed");

        let caught: unknown;
        try {
            assertFileField(data, fileFields);
        } catch (e) {
            caught = e;
        }

        expect(caught).toBeInstanceOf(ClientResponseError);
        expect((caught as ClientResponseError).status).toBe(400);
        expect((caught as ClientResponseError).response).toEqual({ message: "missing_file", expected: fileFields });
    });

    it("rejects an empty body", () => {
        expect(() => assertFileField(new FormData(), fileFields)).toThrow(ClientResponseError);
    });
});

describe("immutable multipart update fields", () => {
    const targets: [Collection, string[]][] = [
        [Collection.trails, ["author"]],
        [Collection.lists, ["author"]],
        [Collection.comments, ["author", "trail"]],
        [Collection.waypoints, ["author", "trail"]],
        [Collection.summit_logs, ["author", "trail"]],
        [Collection.plugin_instances, ["user", "plugin_id"]],
        [Collection.trail_share, ["trail"]],
        [Collection.list_share, ["list"]],
        [Collection.trail_link_share, ["trail"]],
        [Collection.follows, ["follower"]],
        [Collection.settings, ["user"]],
        [Collection.user_category_preferences, ["user"]],
        [Collection.user_subcategory_preferences, ["user"]],
    ];

    it.each(targets)("removes ownership fields and modifiers for %s while preserving content", async (collection, fields) => {
        const data = new FormData();
        for (const field of fields) {
            for (const key of [field, `${field}+`, `${field}-`, `+${field}`, `-${field}`]) {
                data.append(key, "attacker");
                data.append(key, "another-attacker");
            }
        }
        data.set("name", "Authorized edit");
        data.set("text", "Updated content");
        data.set("photos-", "old.jpg");
        data.set("photos+", new Blob(["photo"]), "new.jpg");
        data.set("gpx", new Blob(["track"]), "track.gpx");
        const { event, update } = uploadEvent(data);

        await uploadUpdate(event, collection);

        expect(update).toHaveBeenCalledOnce();
        const [id, forwarded] = update.mock.calls[0];
        expect(id).toBe("record000000001");
        expect([...forwarded.keys()].sort()).toEqual(["gpx", "name", "photos+", "photos-", "text"]);
        expect(forwarded.get("name")).toBe("Authorized edit");
        expect(forwarded.get("text")).toBe("Updated content");
        expect(forwarded.get("photos-")).toBe("old.jpg");
        expect(await forwarded.get("photos+").text()).toBe("photo");
        expect(await forwarded.get("gpx").text()).toBe("track");
    });

    it("sanitizes every multipart JSON payload, including modifiers", async () => {
        const data = new FormData();
        data.set("author", "attacker");
        data.append("@jsonPayload", JSON.stringify({
            author: "attacker", "author+": "attacker", "-author": "owner",
            trail: "other", "trail-": "current", "+trail": "other",
            text: "Authorized edit", "photos-": ["old.jpg"],
        }));
        data.append("@jsonPayload", JSON.stringify({ author: "attacker", trail: "other", name: "Still allowed" }));
        const { event, update } = uploadEvent(data);

        await uploadUpdate(event, Collection.summit_logs);

        const forwarded = update.mock.calls[0][1] as FormData;
        expect(forwarded.has("author")).toBe(false);
        expect(forwarded.getAll("@jsonPayload").map((payload) => JSON.parse(String(payload)))).toEqual([
            { text: "Authorized edit", "photos-": ["old.jpg"] },
            { name: "Still allowed" },
        ]);
    });

    const shares: [Collection, string, Record<string, string>][] = [
        [Collection.trail_share, "trail", { actor: "recipient", permission: "edit" }],
        [Collection.list_share, "list", { actor: "recipient", permission: "view" }],
        [Collection.trail_link_share, "trail", { token: "new-link-token", permission: "view" }],
    ];
    it.each(shares)("keeps recipient and permission changes for %s", async (collection, target, editable) => {
        const data = new FormData();
        data.set(target, "foreign-target");
        for (const [key, value] of Object.entries(editable)) data.set(key, value);
        data.set("@jsonPayload", JSON.stringify({ [target]: "foreign-target", [`${target}+`]: "foreign-target", ...editable }));
        const { event, update } = uploadEvent(data);

        await uploadUpdate(event, collection);

        expect(update).toHaveBeenCalledOnce();
        const forwarded = update.mock.calls[0][1] as FormData;
        expect(forwarded.has(target)).toBe(false);
        for (const [key, value] of Object.entries(editable)) expect(forwarded.get(key)).toBe(value);
        expect(JSON.parse(String(forwarded.get("@jsonPayload")))).toEqual(editable);
    });

    it.each(["{", "null", "[]", "123", '"text"'])("rejects an invalid JSON payload %s before forwarding", async (payload) => {
        const data = new FormData();
        data.set("@jsonPayload", payload);
        const { event, update } = uploadEvent(data);

        await expect(uploadUpdate(event, Collection.trails)).rejects.toMatchObject({ status: 400 });
        expect(update).not.toHaveBeenCalled();
    });

    it("applies the same protection to file-only upload routes", async () => {
        const data = new FormData();
        data.set("photos+", new Blob(["photo"]), "new.jpg");
        data.set("author+", "attacker");
        data.set("trail", "other");
        data.set("@jsonPayload", JSON.stringify({ author: "attacker", "trail-": "current", name: "Edited waypoint" }));
        const { event, update } = uploadEvent(data);

        await upload(event, Collection.waypoints, ["photos"]);

        const forwarded = update.mock.calls[0][1] as FormData;
        expect(forwarded.has("author+")).toBe(false);
        expect(forwarded.has("trail")).toBe(false);
        expect(forwarded.has("photos+")).toBe(true);
        expect(JSON.parse(String(forwarded.get("@jsonPayload")))).toEqual({ name: "Edited waypoint" });
    });

    it("keeps the record ID check before forwarding", async () => {
        const data = new FormData();
        data.set("id", "different000001");
        data.set("name", "Authorized edit");
        const { event, update } = uploadEvent(data);

        await expect(uploadUpdate(event, Collection.trails)).rejects.toMatchObject({
            status: 400, response: { message: "id_mismatch" },
        });
        expect(update).not.toHaveBeenCalled();
    });

    it("keeps required owner and target fields for create requests", async () => {
        const data = new FormData();
        data.set("author", "owner");
        data.set("trail", "target");
        data.set("text", "New log");
        const { event, create, update } = uploadEvent(data);

        await uploadCreate(event, Collection.summit_logs);

        expect(update).not.toHaveBeenCalled();
        expect(create.mock.calls[0][0]).toBe(data);
        expect(data.get("author")).toBe("owner");
        expect(data.get("trail")).toBe("target");
    });
});
