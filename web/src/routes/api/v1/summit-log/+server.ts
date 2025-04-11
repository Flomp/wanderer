import { SummitLogCreateSchema } from '$lib/models/api/summit_log_schema';
import type { SummitLog } from '$lib/models/summit_log';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<SummitLog>(event, Collection.summit_logs);

        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<SummitLog>(event, SummitLogCreateSchema, Collection.summit_logs)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}
