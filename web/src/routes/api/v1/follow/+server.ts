import type { Actor } from '$lib/models/activitypub/actor';
import { FollowCreateSchema } from '$lib/models/api/follow_schema';
import type { Follow } from '$lib/models/follow';
import type { UserAnonymous } from '$lib/models/user';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Follow>(event, Collection.follows);
        for (const follow of r.items) {
            const follower = await event.locals.pb.collection('activitypub_actors').getOne<Actor>(follow.follower, { requestKey: null })
            const followee = await event.locals.pb.collection('activitypub_actors').getOne<Actor>(follow.followee, { requestKey: null })
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
        const data = await event.request.json();
        const safeData = FollowCreateSchema.parse(data);

        const followerActor = event.locals.pb.collection("activitypub_actors").getFirstListItem(`user = '${event.locals.user.id}'`)
        const followeeActor = event.locals.pb.collection("activitypub_actors").getFirstListItem(`user = '${safeData.followee}'`)

        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}