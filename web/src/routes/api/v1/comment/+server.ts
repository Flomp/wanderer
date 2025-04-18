import { CommentCreateSchema } from '$lib/models/api/comment_schema';
import type { Comment } from '$lib/models/comment';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Comment>(event, Collection.comments);
        for (const comment of r.items) {
            if (!comment.expand?.author) {
                comment.expand = {
                    author: await event.locals.pb.collection('users_anonymous').getOne(comment.author)
                }
            }
        }
        return json(r)
    } catch (e) {
        throw handleError(e)
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<Comment>(event, CommentCreateSchema, Collection.comments)
        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}