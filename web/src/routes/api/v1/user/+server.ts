import { env } from '$env/dynamic/public';
import type { User } from '$lib/models/user';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit'
import { ClientResponseError } from 'pocketbase';

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


export async function PUT(event: RequestEvent) {
    const data = await event.request.json()
    try {
        if (env.PUBLIC_DISABLE_SIGNUP === "true") {
            throw new ClientResponseError({ status: 401, response: { messgage: "Forbidden" } })
        }
        const r = await pb.collection('users').create<User>(data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

