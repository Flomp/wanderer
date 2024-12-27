import type { Waypoint } from "$lib/models/waypoint";
import { Collection, handleError, upload } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    try {
        const r = await upload<Waypoint>(event, Collection.waypoints);
        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }
}