import { env } from '$env/dynamic/public';
import { pb } from '$lib/pocketbase';
import type { User } from '$lib/stores/user_store';
import { error, json, type RequestEvent } from '@sveltejs/kit'
import { ClientResponseError } from 'pocketbase';

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

