import { Collection, handleError, list } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import { type Actor } from "$lib/models/activitypub/actor"

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Actor>(event, Collection.activitypub_actors);
        return json(r)
    } catch (e) {
        return handleError(e)
    }
}


