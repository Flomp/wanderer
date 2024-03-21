import type { Waypoint } from '$lib/models/waypoint';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r: Waypoint[] = await pb.collection('waypoints').getFullList<Waypoint>()
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function PUT(event: RequestEvent) {
    const data = await event.request.json();

    try {
        const r = await pb.collection('waypoints').create<Waypoint>(data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}