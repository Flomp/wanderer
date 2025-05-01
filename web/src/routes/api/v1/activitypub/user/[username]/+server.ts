import { handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';


export async function GET(event: RequestEvent) {

    try {
        const r = await event.locals.pb.send(`/activitypub/user/${event.params.username}`, { method: "GET", fetch: event.fetch })

        return json(r);
    } catch (e) {
        throw handleError(e)
    }


}

