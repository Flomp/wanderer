import { TrailRecommendSchema } from '$lib/models/api/trail_schema';
import type { Trail } from '$lib/models/trail';
import { handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const searchParams = Object.fromEntries(event.url.searchParams);
        const safeSearchParams = TrailRecommendSchema.parse(searchParams);
        
        const r = await event.locals.pb.send("/trail/recommend?" + new URLSearchParams({
            size: safeSearchParams.size?.toString() ?? ""
        }), {
            method: "GET",
        });

        const result: Trail[] = r;

        for (const t of result) {
            if (!t.author || !event.locals.pb.authStore.record) {
                continue;
            }
            if (!t.expand) {
                t.expand = {} as any
            }
            t.expand!.author = await event.locals.pb.collection("users_anonymous").getOne(t.author);
        }

        return json(result)
    } catch (e: any) {
        throw handleError(e);
    }
}
