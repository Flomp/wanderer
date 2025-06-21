import type { List } from '$lib/models/list';
import { Collection, handleError, uploadCreate } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function PUT(event: RequestEvent) {
    try {
        const r = await uploadCreate<List>(event, Collection.lists)
        return json(r);
    } catch (e) {
        return handleError(e)
    }
}