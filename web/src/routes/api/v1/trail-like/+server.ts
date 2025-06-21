import { TrailLikeCreateSchema } from '$lib/models/api/trail_like_schema';
import type { TrailLike } from '$lib/models/trail_like';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<TrailLike>(event, Collection.trail_share);
        return json(r)
    } catch (e: any) {
        handleError(e)
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<TrailLike>(event, TrailLikeCreateSchema, Collection.trail_like)

        return json(r);
    } catch (e) {
        return handleError(e)
    }
}