
import { handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';


export async function GET(event: RequestEvent) {
    const id = event.params.id;

    try {
        const comment = await event.locals.pb.send("/activitypub/comment/" + id, {
            method: "GET",
            fetch: event.fetch,
        })

        const headers = new Headers()
        headers.append("Content-Type", "application/activity+json")

        return json({
            "@context": [
                "https://www.w3.org/ns/activitystreams",
            ],
            ...comment
        }, { status: 200, headers });
    } catch (e) {
        return handleError(e)
    }

}