import { TrailCreateSchema } from '$lib/models/api/trail_schema';
import type { Trail } from '$lib/models/trail';
import { pb } from '$lib/pocketbase';
import { Collection, create, handleError, list } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Trail>(event, Collection.trails);

        for (const t of r.items) {
            if (!t.author || !pb.authStore.model) {
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
        const r = await create<Trail>(event, TrailCreateSchema, Collection.trails)
        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}