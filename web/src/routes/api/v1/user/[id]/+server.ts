import type { User } from '$lib/models/user';
import { pb } from '$lib/pocketbase';
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
        if (data.email != pb.authStore.model!.email) {
            const r = await pb.collection('users').requestEmailChange(data.email);
            pb.authStore.model!.email = data.email;
        }
        const r = await pb.collection('users').update<User>(event.params.id as string, data)
        if (data.password) {
            const r = await pb.collection('users').authWithPassword(data.email ?? data.username!, data.password);
            return json(r.record)
        } else {
            return json(r);
        }
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
