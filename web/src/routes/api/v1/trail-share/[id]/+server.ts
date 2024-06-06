import type { TrailShare } from "$lib/models/trail_share";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await pb.collection('trail_share')
            .getOne<TrailShare>(event.params.id as string)
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function POST(event: RequestEvent) {
    const data = await event.request.json()

    try {
        const r = await pb.collection('trail_share').update<TrailShare>(event.params.id as string, data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await pb.collection('trail_share').delete(event.params.id as string)
        return json({ 'acknowledged': r });
    } catch (e: any) {
        throw error(e.status, e)
    }
}
