import { pb } from "$lib/pocketbase";
import { error, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    const parts = [];
    parts.push("api");
    parts.push("files");
    parts.push(encodeURIComponent(event.params.collection as string));
    parts.push(encodeURIComponent(event.params.record as string));
    parts.push(encodeURIComponent(event.params.file as string));

    let fileURL = pb.buildUrl(parts.join("/"));

    try {
        const r = await event.fetch(fileURL)
        return new Response(r.body, {headers: r.headers});
    } catch (e: any) {
        console.log(e);
        
        throw error(500, e);
    }

}