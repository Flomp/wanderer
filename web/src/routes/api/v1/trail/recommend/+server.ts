import { TrailRecommendSchema } from '$lib/models/api/trail_schema';
import type { Trail, TrailSearchResult } from '$lib/models/trail';
import { handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';
import type { Hits } from 'meilisearch';

export async function GET(event: RequestEvent) {
    try {
        const searchParams = Object.fromEntries(event.url.searchParams);
        const safeSearchParams = TrailRecommendSchema.parse(searchParams);

        const numberOfTrails = (await event.locals.ms.index("trails").search("", {limit: 1})).estimatedTotalHits
        const randomOffset = (safeSearchParams.size ?? 0) > numberOfTrails ? 0 : Math.floor(Math.random() * (numberOfTrails - 1) + 1)
        const response = await event.locals.ms.index("trails").search("", {limit: safeSearchParams.size, offset: randomOffset})

        return json(response.hits)
    } catch (e: any) {
        return handleError(e);
    }
}
