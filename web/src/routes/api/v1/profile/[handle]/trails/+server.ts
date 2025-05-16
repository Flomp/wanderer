import type { TrailSearchResult } from '$lib/models/trail';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import type { SearchResponse } from 'meilisearch';
import { ClientResponseError } from 'pocketbase';

export async function POST(event: RequestEvent) {
    const handle = event.params.handle;
    if (!handle) {
        return error(400, { message: "Bad request" })
    }

    try {
        const actor = await event.locals.pb.send(`/activitypub/actor?resource=acct:${handle}`, { method: "GET", fetch: event.fetch, });

        const data = await event.request.json()

        let r: SearchResponse<TrailSearchResult>;
        if (actor.isLocal) {
            r = await event.locals.ms.index("trails").search(data.q, { ...data.options, filter: `author = ${actor.id}` });
        } else {
            const origin = new URL(actor.iri).origin
            const url = `${origin}/api/v1/profile/${actor.username}/trails?` + event.url.searchParams
            const response = await event.fetch(url, { method: 'POST', body: JSON.stringify(data) })

            if (!response.ok) {
                const errorResponse = await response.json()
                throw new ClientResponseError({ status: response.status, response: errorResponse });
            }
            r = await response.json()

            r.hits.forEach(h => {
                h.thumbnail = `${origin}/api/v1/files/trails/${h.id}/${h.thumbnail}`;
                h.domain = actor.domain
            })
        }


        return json(r)
    } catch (e) {
        return handleError(e)
    }
}
