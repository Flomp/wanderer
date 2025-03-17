import { TrailRecommendSchema } from '$lib/models/api/trail_schema';
import type { Trail } from '$lib/models/trail';
import { pb } from '$lib/pocketbase';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import { ClientResponseError } from 'pocketbase';

export async function GET(event: RequestEvent) {
    try {
        const searchParams = Object.fromEntries(event.url.searchParams);
        const safeSearchParams = TrailRecommendSchema.parse(searchParams);

        const r = await event.fetch(pb.buildURL("/trail/recommend?" + new URLSearchParams({
            size: safeSearchParams.size?.toString() ?? ""
        })));

        if (!r.ok) {
            const response = await r.json()
            console.error(response)
            throw new ClientResponseError({ status: response.code, response })
        }

        const result: Trail[] = await r.json();

        for (const t of result) {
            if (!t.author || !pb.authStore.record) {
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
