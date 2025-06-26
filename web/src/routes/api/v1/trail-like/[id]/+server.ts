import type { TrailLike } from "$lib/models/trail_like";
import { Collection, handleError, remove, show } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await show<TrailLike>(event, Collection.trail_like)
        return json(r)
    } catch (e: any) {
        return handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.trail_like)
        return json(r);
    } catch (e: any) {
        return handleError(e)
    }
}
