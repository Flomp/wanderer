import type { TrailSearchResult } from '$lib/models/trail';
import { actorFromRemote, splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import type { SearchResponse } from 'meilisearch';
import { ClientResponseError } from 'pocketbase';

export async function POST(event: RequestEvent) {
    const fullUsername = event.params.username;
    if (!fullUsername) {
        return error(400, { message: "Bad request" })
    }
    const [username, domain] = splitUsername(fullUsername)


    try {
        const { actor: fetchedActor, remote: remoteActor } = await actorFromRemote(domain, username, event.fetch);

        try {
            const dbActor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`iri='${fetchedActor.iri}'`)
            fetchedActor.id = dbActor.id
            fetchedActor.isLocal = dbActor.isLocal
        } catch (e) {
            const dbActor = await event.locals.pb.collection("activitypub_actors").create(fetchedActor);
            fetchedActor.id = dbActor.id
        }

        const data = await event.request.json()

        let r: SearchResponse<TrailSearchResult>;
        if (fetchedActor.isLocal) {
            r = await event.locals.ms.index("trails").search(data.q, { ...data.options, filter: `author = ${fetchedActor.id}` });
        } else {
            const origin = new URL(fetchedActor.iri).origin
            const url = origin + '/api/v1/profile/' + username + '/trails'
            const response = await event.fetch(url, { method: 'POST', body: JSON.stringify(data) })

            if (!response.ok) {
                const errorResponse = await response.json()
                throw new ClientResponseError({ status: 500, response: errorResponse });
            }
            r = await response.json()

            r.hits.forEach(h => {
                h.thumbnail = `${origin}/api/v1/files/trails/${h.id}/${h.thumbnail}`;
                h.domain = fetchedActor.domain
            })
        }


        return json(r)
    } catch (e) {
        return handleError(e)
    }
}
