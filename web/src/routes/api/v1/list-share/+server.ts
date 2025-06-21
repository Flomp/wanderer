import type { Actor } from '$lib/models/activitypub/actor';
import { ListShareCreateSchema } from '$lib/models/api/list_share_schema';
import type { ListShare } from '$lib/models/list_share';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<ListShare>(event, Collection.list_share);
        return json(r)
    } catch (e: any) {
        return handleError(e);
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const data = await event.request.json();
        const safeData = ListShareCreateSchema.parse(data);

        const { actor }: { actor: Actor } = await event.locals.pb.send(`/activitypub/actor?iri=${safeData.actor}`, { method: "GET", fetch: event.fetch, });
        safeData.actor = actor.id!;

        const r = await event.locals.pb.collection(Collection.list_share).create<ListShare>(safeData)

        return json(r);
    } catch (e) {
        return handleError(e)
    }
}