import type { Comment } from '$lib/models/comment';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    const filter = event.url.searchParams.get('filter') ?? ""

    try {
        const r: Comment[] = await pb.collection('comments').getFullList<Comment>({
            expand: "author",
            filter: filter,
            sort: "-created",
        })

        for (const comment of r) {
            if(!comment.expand?.author) {
                comment.expand = {
                    author: await pb.collection('users_anonymous').getOne(comment.author)
                }
            }
        }
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