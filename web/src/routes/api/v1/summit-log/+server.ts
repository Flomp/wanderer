import { SummitLogCreateSchema } from '$lib/models/api/summit_log_schema';
import type { SummitLog } from '$lib/models/summit_log';
import { pb } from '$lib/pocketbase';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<SummitLog>(event, Collection.summit_logs);

        for (const t of r.items) {
            if (!t.author || !pb.authStore.record) {
                continue;
            }
            if (!t.expand) {
                t.expand = {} as any
            }
            t.expand!.author = await pb.collection("users_anonymous").getOne(t.author);
        }
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
