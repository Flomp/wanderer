import { pb } from '$lib/pocketbase';
import type { User } from '$lib/stores/user_store';
import { error, json, type RequestEvent } from '@sveltejs/kit'

export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const r = await pb.collection('lists').update<User>(data.id, data.user)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

export async function PUT(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const r = await pb.collection('users').create<User>(data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

export async function DELETE(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const r = await pb.collection('users').delete(data.id)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}
