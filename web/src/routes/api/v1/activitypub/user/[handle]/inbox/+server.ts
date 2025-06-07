import { handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';
import type { APActivity } from 'activitypub-types';

export async function POST(event: RequestEvent) {

    try {
        const activity: APActivity = await event.request.json()
        if (!activity.actor) {
            return json("Bad request", { status: 400 });
        }

        const success = await event.locals.pb.send("/activitypub/activity/process", {
            method: "POST", 
            fetch: event.fetch,
            headers: {
                'X-Forwarded-Path': event.url.pathname,
                'Content-Type': event.request.headers.get("content-type")!,
                signature: event.request.headers.get("signature")!,
                date: event.request.headers.get("date")!,
                digest: event.request.headers.get("digest")!
            },
            body: JSON.stringify(activity)
        })

        if (success === false) {
            return json("Invalid header signature", { status: 400 });
        }

        const headers = new Headers()
        headers.append("Content-Type", "application/activity+json")

        return json("", { status: 200, headers });
    } catch (e) {
        return handleError(e)
    }


}

