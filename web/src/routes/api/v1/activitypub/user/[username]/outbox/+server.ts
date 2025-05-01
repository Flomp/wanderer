// src/routes/inbox/+server.ts
import { handleError } from '$lib/util/api_util';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from './$types';

export async function GET(event: RequestEvent) {
    try {
        const r = await event.locals.pb.send(`/activitypub/user/${event.params.username}/outbox?` + event.url.searchParams, { method: "GET", fetch: event.fetch })

        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}

