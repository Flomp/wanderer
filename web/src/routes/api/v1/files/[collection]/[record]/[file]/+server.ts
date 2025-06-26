import { error, json, type RequestEvent } from "@sveltejs/kit";
import { z } from "zod";

export async function GET(event: RequestEvent) {

    const safeParams = z.object({
        collection: z.string(),
        record: z.string().length(15),
        file: z.string()
    }).safeParse(event.params)

    if (!safeParams.success) {
        throw json({ message: safeParams.error }, { status: 400 })
    }

    const safeSearchParams = z.object({
        thumb: z.string().regex(/[0-9]*x[0-9]*[tbf]?/).optional()
    }).safeParse(Object.fromEntries(event.url.searchParams))

    if (!safeSearchParams.success) {
        throw json({ message: safeSearchParams.error }, { status: 400 })
    }

    const parts = [];
    parts.push("api");
    parts.push("files");
    parts.push(encodeURIComponent(safeParams.data.collection));
    parts.push(encodeURIComponent(safeParams.data.record));
    parts.push(encodeURIComponent(safeParams.data.file));

    let fileURL = event.locals.pb.buildURL(parts.join("/") + '?' + new URLSearchParams(safeSearchParams.data));

    try {
        const r = await event.fetch(fileURL)
        return new Response(r.body, { headers: r.headers });
    } catch (e: any) {
        throw error(500, e);
    }

}