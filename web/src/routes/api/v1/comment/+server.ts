import type { Comment } from '$lib/models/comment';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r: Comment[] = await pb.collection('comments').getFullList<Comment>({
            expand: "author",
            sort: "-created",
        })
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function PUT(event: RequestEvent) {
    const data = await event.request.json();

    try {
        const r = await pb.collection('comments').create<Comment>(data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}