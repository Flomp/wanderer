import { ListShareCreateSchema } from '$lib/models/api/list_share_schema';
import type { ListShare } from '$lib/models/list_share';
import type { User } from '$lib/models/user';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<ListShare>(event, Collection.list_share);

        for (const share of r.items) {
            const anonymous_user = await event.locals.pb.collection('users_anonymous').getOne<User>(share.user)
            share.expand = {
                user: anonymous_user
            }
        }
        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<ListShare>(event, ListShareCreateSchema, Collection.list_share)
        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}