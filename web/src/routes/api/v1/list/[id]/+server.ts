import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";
import type { List } from "postcss/lib/list";

export async function GET(event: RequestEvent) {
    try {
        const r = await pb.collection('lists')
            .getOne<List>(event.params.id as string, { expand: "trails,trails.waypoints,trails.category" })
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function POST(event: RequestEvent) {
    const data = await event.request.json()

    try {
        const r = await pb.collection('lists').update<List>(event.params.id as string, data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await pb.collection('lists').delete(event.params.id as string)
        return json({ 'acknowledged': r });
    } catch (e: any) {
        throw error(e.status, e)
    }
}
