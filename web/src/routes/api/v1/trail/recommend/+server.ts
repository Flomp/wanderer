import { TrailRecommendSchema } from '$lib/models/api/trail_schema';
import type { Trail } from '$lib/models/trail';
import { pb } from '$lib/pocketbase';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const searchParams = Object.fromEntries(event.url.searchParams);
        const safeSearchParams = TrailRecommendSchema.parse(searchParams);

        const r = await event.fetch(pb.buildUrl("/trail/recommend?" + new URLSearchParams({
            size: safeSearchParams.size?.toString() ?? ""
        })));

        const result: Trail[] = await r.json();

        for (const t of result) {
            if (!t.author || !pb.authStore.model) {
                continue;
            }
            if (!t.expand) {
                t.expand = {} as any
            }
            t.expand!.author = await pb.collection("users_anonymous").getOne(t.author);
        }

        return json(result)
    } catch (e: any) {
        throw handleError(e);
    }
}
