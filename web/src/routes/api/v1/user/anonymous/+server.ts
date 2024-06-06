import type { User } from '$lib/models/user';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    const page = event.url.searchParams.get("page") ?? "0";
    const perPage = event.url.searchParams.get("per-page") ?? "10";
    const sort = event.url.searchParams.get("sort") ?? ""
    const filter = event.url.searchParams.get("filter") ?? "";

    try {
        const r = await pb.collection('users_anonymous')
            .getList<User>(parseInt(page), parseInt(perPage), { sort: sort ?? "", filter: filter ?? "" })
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

