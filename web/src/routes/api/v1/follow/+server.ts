import type { Follow } from '$lib/models/follow';
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
                items: await pb.collection('follows')
                    .getFullList<Follow>({ sort: sort, filter: filter, requestKey: filter })
            }
        } else {
            r = await pb.collection('follows')
                .getList<Follow>(parseInt(page), parseInt(perPage), { sort: sort ?? "", filter: filter ?? "", requestKey: filter })
        }
        for (const follow of r.items) {
            const follower = await pb.collection('users_anonymous').getOne<UserAnonymous>(follow.follower, { requestKey: filter })
            const followee = await pb.collection('users_anonymous').getOne<UserAnonymous>(follow.followee, { requestKey: filter })
            follow.expand = {
                follower, followee
            }
        }
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function PUT(event: RequestEvent) {
    const data = await event.request.json();

    try {
        const r = await pb.collection('follows').create<Follow>(data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}