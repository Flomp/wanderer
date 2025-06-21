import type { Actor } from '$lib/models/activitypub/actor';
import { RecordListOptionsSchema } from '$lib/models/api/base_schema';
import { type TimelineItem } from '$lib/models/timeline';
import { Collection, handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import { ClientResponseError, type ListResult } from 'pocketbase';

export async function GET(event: RequestEvent) {
    const handle = event.params.handle;
    if (!handle) {
        return error(400, { message: "Bad request" })
    }

    try {
        const {actor, error} = await event.locals.pb.send(`/activitypub/actor?resource=acct:${handle}`, { method: "GET", fetch: event.fetch, });

        const searchParams = Object.fromEntries(event.url.searchParams);
        const safeSearchParams = RecordListOptionsSchema.parse(searchParams);

        let timeline: ListResult<TimelineItem>;
        if (actor.isLocal) {
            timeline = await event.locals.pb.collection(Collection.timeline)
                .getList<TimelineItem>(safeSearchParams.page, safeSearchParams.perPage, { ...safeSearchParams, filter: `author='${actor.iri}'` })
        } else {
            const origin = new URL(actor.iri).origin
            const timelineURL = `${origin}/api/v1/profile/${actor.username}/timeline?` + event.url.searchParams

            const response = await event.fetch(timelineURL, { method: 'GET' })
            if (!response.ok) {
                const errorResponse = await response.json()
                throw new ClientResponseError({ status: response.status, response: errorResponse });
            }
            timeline = await response.json()

            timeline.items.forEach(i => {
                i.photos = i. photos.map(p =>
                    `${origin}/api/v1/files/${i.type}s/${i.id}/${p}`
                )
            })
        }


        return json(timeline)
    } catch (e) {
        console.error(e)
        return handleError(e)
    }
}
