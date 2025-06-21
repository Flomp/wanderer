import { WaypointUpdateSchema } from "$lib/models/api/waypoint_schema";
import type { Waypoint } from "$lib/models/waypoint";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await show<Waypoint>(event, Collection.waypoints)
        return json(r)
    } catch (e: any) {
        return handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<Waypoint>(event, WaypointUpdateSchema, Collection.waypoints)
        return json(r);
    } catch (e: any) {
        return handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.waypoints)
        return json(r);
    } catch (e: any) {
        return handleError(e)
    }
}
