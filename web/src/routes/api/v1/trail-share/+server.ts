import type { Actor } from '$lib/models/activitypub/actor';
import { TrailShareCreateSchema } from '$lib/models/api/trail_share_schema';
import type { TrailShare } from '$lib/models/trail_share';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<TrailShare>(event, Collection.trail_share);
        return json(r)
    } catch (e: any) {
        handleError(e)
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const data = await event.request.json();
        const safeData = TrailShareCreateSchema.parse(data);

        const { actor }: { actor: Actor } = await event.locals.pb.send(`/activitypub/actor?iri=${safeData.actor}`, { method: "GET", fetch: event.fetch, });
        safeData.actor = actor.id!;

        const r = await event.locals.pb.collection(Collection.trail_share).create<TrailShare>(safeData)

        return json(r);
    } catch (e) {
        return handleError(e)
    }
}