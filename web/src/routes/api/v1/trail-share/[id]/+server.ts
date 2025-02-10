import { TrailShareUpdateSchema } from "$lib/models/api/trail_share_schema";
import type { TrailShare } from "$lib/models/trail_share";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await show<TrailShare>(event, Collection.trail_share)
        return json(r)
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<TrailShare>(event, TrailShareUpdateSchema, Collection.trail_share)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.trail_share)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}
