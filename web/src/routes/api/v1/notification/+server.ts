import type { Notification } from '$lib/models/notification';
import type { UserAnonymous } from '$lib/models/user';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    const page = event.url.searchParams.get("page") ?? "0";
    const perPage = event.url.searchParams.get("per-page") ?? "10";
    const sort = event.url.searchParams.get('sort') ?? ""
    const filter = event.url.searchParams.get("filter") ?? "";

    try {
        let r;
        if (parseInt(perPage) < 0) {
            r = {
                items: await pb.collection('notifications')
                    .getFullList<Notification>({ sort: sort, filter: filter, requestKey: filter })
            }
        } else {
            r = await pb.collection('notifications')
                .getList<Notification>(parseInt(page), parseInt(perPage), { sort: sort ?? "", filter: filter ?? "", requestKey: filter })
        }
        for (const notification of r.items) {
            const recipient = await pb.collection('users_anonymous').getOne<UserAnonymous>(notification.recipient, { requestKey: filter })
            const author = await pb.collection('users_anonymous').getOne<UserAnonymous>(notification.author, { requestKey: filter })
            notification.expand = {
                recipient, author
            }
        }
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}
