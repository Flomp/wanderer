import { pb } from '$lib/pocketbase';
import type { User } from '$lib/stores/user_store';
import { error, json, type RequestEvent } from '@sveltejs/kit'

export async function PUT(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const r = await pb.collection('users').create<User>(data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

