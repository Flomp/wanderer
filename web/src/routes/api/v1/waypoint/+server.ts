import { WaypointCreateSchema } from '$lib/models/api/waypoint_schema';
import type { Waypoint } from '$lib/models/waypoint';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Waypoint>(event, Collection.waypoints);
        return json(r)
    } catch (e: any) {
        return handleError(e);
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<Waypoint>(event, WaypointCreateSchema, Collection.waypoints)
        return json(r);
    } catch (e) {
        return handleError(e)
    }
}