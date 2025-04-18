import type { UserAnonymous } from '$lib/models/user';
import { Collection, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    let filter = event.url.searchParams.get("filter");

    if (!filter) {
        filter = "private=false"
    } else {
        filter += "&&private=false"
    }
    event.url.searchParams.set("filter", filter);

    try {
        const r = await list<UserAnonymous>(event, Collection.users_anonymous);

        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}

