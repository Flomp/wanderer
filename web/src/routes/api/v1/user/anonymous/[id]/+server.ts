import type { User } from '$lib/models/user';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit'

export async function GET(event: RequestEvent) {
    try {
        const r = await pb.collection('users_anonymous')
            .getOne<User>(event.params.id as string)
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}