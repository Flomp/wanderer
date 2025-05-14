import { RecordListOptionsSchema } from '$lib/models/api/base_schema';
import { type TimelineItem } from '$lib/models/timeline';
import { actorFromRemote, splitUsername } from '$lib/util/activitypub_util';
import { Collection, handleError, list } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import { ClientResponseError, type ListResult } from 'pocketbase';

export async function GET(event: RequestEvent) {
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

        const searchParams = Object.fromEntries(event.url.searchParams);
        const safeSearchParams = RecordListOptionsSchema.parse(searchParams);

        let timeline: ListResult<TimelineItem>;
        if (fetchedActor.isLocal) {
            timeline = await event.locals.pb.collection(Collection.timeline)
                .getList<TimelineItem>(safeSearchParams.page, safeSearchParams.perPage, { ...safeSearchParams, filter: `author='${fetchedActor.iri}'` })
        } else {
            const timelineURL = new URL(fetchedActor.iri).origin + '/api/v1/profile/@' + username + '/timeline'
            const response = await event.fetch(timelineURL, { method: 'GET' })
            if (!response.ok) {
                const errorResponse = await response.json()
                throw new ClientResponseError({ status: 500, response: errorResponse });
            }
            timeline = await response.json()
        }


        return json(timeline)
    } catch (e) {
        return handleError(e)
    }
}
