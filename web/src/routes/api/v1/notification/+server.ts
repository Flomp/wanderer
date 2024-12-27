import type { Notification } from '$lib/models/notification';
import type { UserAnonymous } from '$lib/models/user';
import { pb } from '$lib/pocketbase';
import { Collection, list } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Notification>(event, Collection.notifications);

        for (const notification of r.items) {
            const recipient = await pb.collection('users_anonymous').getOne<UserAnonymous>(notification.recipient, { requestKey: null })
            const author = await pb.collection('users_anonymous').getOne<UserAnonymous>(notification.author, { requestKey: null })
            notification.expand = {
                recipient, author
            }
        }
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}