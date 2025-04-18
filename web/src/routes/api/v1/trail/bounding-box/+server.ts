import { type TrailBoundingBox, type TrailFilterValues } from '$lib/models/trail';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    if (!event.locals.pb.authStore.record) {
        return json({
            max_lat: 0,
            min_lat: 0,
            max_lon: 0,
            min_lon: 0
        });
    }
    try {
        const r = await event.locals.pb.collection('trails_bounding_box').getOne<TrailBoundingBox>(event.locals.pb.authStore.record!.id)
        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}
