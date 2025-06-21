import { TrailLikeCreateSchema } from '$lib/models/api/trail_like_schema';
import { Collection, handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';


// this route exists so we can delete a like without knowing its ID
export async function POST(event: RequestEvent) {
    try {

        const data = await event.request.json();
        const safeData = TrailLikeCreateSchema.parse(data);

        const like = await event.locals.pb.collection(Collection.trail_like).getFirstListItem(`actor='${safeData.actor}'&&trail='${safeData.trail}'`)

        const r = await event.locals.pb.collection(Collection.trail_like).delete(like.id)

        return json({ 'acknowledged': r })
    } catch (e: any) {
        handleError(e)
    }
}