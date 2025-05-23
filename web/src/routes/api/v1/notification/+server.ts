import type { Notification } from '$lib/models/notification';
import { Collection, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Notification>(event, Collection.notifications);

        return json(r)
    } catch (e: any) {
        return handleError(e);
    }
}
