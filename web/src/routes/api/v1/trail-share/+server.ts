import { TrailShareCreateSchema } from '$lib/models/api/trail_share_schema';
import type { TrailShare } from '$lib/models/trail_share';
import type { UserAnonymous } from '$lib/models/user';
import { Collection, handleError, list, create } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<TrailShare>(event, Collection.trail_share);
        for (const share of r.items) {
            const anonymous_user = await event.locals.pb.collection('users_anonymous').getOne<UserAnonymous>(share.user)
            share.expand = {
                user: anonymous_user
            }
        }
        return json(r)
    } catch (e: any) {
        handleError(e)
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<TrailShare>(event, TrailShareCreateSchema, Collection.trail_share)
        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}