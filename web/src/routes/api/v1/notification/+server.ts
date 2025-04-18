import type { Notification } from '$lib/models/notification';
import type { UserAnonymous } from '$lib/models/user';
import { Collection, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Notification>(event, Collection.notifications);

        for (const notification of r.items) {
            const recipient = await event.locals.pb.collection('users_anonymous').getOne<UserAnonymous>(notification.recipient, { requestKey: null })
            const author = await event.locals.pb.collection('users_anonymous').getOne<UserAnonymous>(notification.author, { requestKey: null })
            notification.expand = {
                recipient, author
            }
        }
        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}
