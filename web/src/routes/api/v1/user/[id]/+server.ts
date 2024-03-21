import { pb } from '$lib/pocketbase';
import type { User } from '$lib/stores/user_store';
import { error, json, type RequestEvent } from '@sveltejs/kit'

export async function GET(event: RequestEvent) {
    try {
        const r = await pb.collection('users')
            .getOne<User>(event.params.id as string)
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const r = await pb.collection('users').update<User>(event.params.id as string, data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await pb.collection('users').delete(event.params.id as string)
        return json({ 'acknowledged': r });
    } catch (e: any) {
        throw error(e.status, e)
    }
}
