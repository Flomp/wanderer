import { CommentUpdateSchema } from "$lib/models/api/comment_schema";
import type { Comment } from "$lib/models/comment";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await show<Comment>(event, Collection.comments)
        if (!r.expand) {
            r.expand = {} as any
        }
        r.expand!.author = await event.locals.pb.collection('users_anonymous').getOne(r.author)

        return json(r)
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<Comment>(event, CommentUpdateSchema, Collection.comments)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.comments)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}
