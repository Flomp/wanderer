import type { SummitLog } from '$lib/models/summit_log';
import { Collection, handleError, uploadCreate } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function PUT(event: RequestEvent) {
    try {
        const r = await uploadCreate<SummitLog>(event, Collection.summit_logs)
        r.date = r.date?.substring(0, 10) ?? "";

        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}
