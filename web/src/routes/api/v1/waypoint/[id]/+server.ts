import type { Waypoint } from "$lib/models/waypoint";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await pb.collection('waypoints')
            .getOne<Waypoint>(event.params.id as string)
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function POST(event: RequestEvent){
    const data = await event.request.json()
    try {
        const r = await await pb.collection("waypoints").update<Waypoint>(event.params.id as string, data);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await pb.collection('waypoints').delete(event.params.id as string)
        return json({ 'acknowledged': r });
    } catch (e: any) {
        throw error(e.status, e)
    }
}
