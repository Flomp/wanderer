import { type TrailFilterValues } from '$lib/models/trail';
import { handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    if (!event.locals.pb.authStore.record) {
        return json({
            min_distance: 0,
            max_distance: 20000,
            min_elevation_gain: 0,
            max_elevation_gain: 4000,
            min_elevation_loss: 0,
            max_elevation_loss: 4000
        });
    }
    try {
        const r = await event.locals.pb.collection('trails_filter').getOne<TrailFilterValues>(event.locals.pb.authStore.record!.id)
        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}
