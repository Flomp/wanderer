import { TagCreateSchema } from '$lib/models/api/tag_schema';
import type { Tag } from '$lib/models/tag';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Tag>(event, Collection.tags);

        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<Tag>(event, TagCreateSchema, Collection.tags)
        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}