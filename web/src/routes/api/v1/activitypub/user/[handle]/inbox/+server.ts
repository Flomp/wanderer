import { env } from '$env/dynamic/private';
import type { Actor } from '$lib/models/activitypub/actor';
import { splitUsername } from '$lib/util/activitypub_util';
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
            method: "POST", fetch: event.fetch,
            headers: {
                'X-Forwarded-Path': event.url.pathname,
                signature: event.request.headers.get("signature")!,
                date: event.request.headers.get("date")!,
                digest: event.request.headers.get("digest")!
            },
            body: activity
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

