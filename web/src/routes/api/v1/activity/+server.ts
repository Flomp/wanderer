import type { Activity } from "$lib/models/activity";
import { Collection, handleError, list } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Activity>(event, Collection.activities);
        return json(r)
    } catch (e) {
        throw handleError(e)
    }
}

