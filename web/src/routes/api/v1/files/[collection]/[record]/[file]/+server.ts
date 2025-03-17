import { pb } from "$lib/pocketbase";
import { error, type RequestEvent } from "@sveltejs/kit";
import { z } from "zod";

export async function GET(event: RequestEvent) {

    const safeParams = z.object({
        collection: z.string(),
        record: z.string().length(15),
        file: z.string()
    }).safeParse(event.params)

    if (!safeParams.success) {
        throw error(400, safeParams.error)
    }

    const parts = [];
    parts.push("api");
    parts.push("files");
    parts.push(encodeURIComponent(safeParams.data.collection));
    parts.push(encodeURIComponent(safeParams.data.record));
    parts.push(encodeURIComponent(safeParams.data.file));

    let fileURL = pb.buildURL(parts.join("/"));

    try {
        const r = await event.fetch(fileURL)
        return new Response(r.body, { headers: r.headers });
    } catch (e: any) {
        throw error(500, e);
    }

}