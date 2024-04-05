import type { Trail } from "$lib/models/trail";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    const expand = event.url.searchParams.get("expand") ?? ""

    try {
        const r = await pb.collection('trails')
            .getOne<Trail>(event.params.id as string, { expand: expand ?? "" })

        // remove time from dates
        r.date = r.date?.substring(0, 10) ?? ""
        for (const log of r.expand?.summit_logs ?? []) {
            log.date = log.date.substring(0, 10)
        }
        return json(r)
    } catch (e: any) {       
        throw error(e.status || 500, e);
    }
}


export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const r = await pb.collection("trails").update<Trail>(event.params.id as string, data, { expand: "category,waypoints,summit_logs" });
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await pb.collection('trails').delete(event.params.id as string)
        return json({ 'acknowledged': r });
    } catch (e: any) {
        throw error(e.status, e)
    }
}
