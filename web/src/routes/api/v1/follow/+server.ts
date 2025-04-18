import { FollowCreateSchema } from '$lib/models/api/follow_schema';
import type { Follow } from '$lib/models/follow';
import type { UserAnonymous } from '$lib/models/user';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Follow>(event, Collection.follows);
        for (const follow of r.items) {
            const follower = await event.locals.pb.collection('users_anonymous').getOne<UserAnonymous>(follow.follower, { requestKey: null })
            const followee = await event.locals.pb.collection('users_anonymous').getOne<UserAnonymous>(follow.followee, { requestKey: null })
            follow.expand = {
                follower, followee
            }
        }
        return json(r)
    } catch (e) {
        throw handleError(e)
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<Follow>(event, FollowCreateSchema, Collection.follows)
        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}