import { type TrailBoundingBox, type TrailFilterValues } from '$lib/models/trail';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    if (!pb.authStore.model) {
        return json({
            max_lat: 0,
            min_lat: 0,
            max_lon: 0,
            min_lon: 0
        });
    }
    try {
        const r = await pb.collection('trails_bounding_box').getOne<TrailBoundingBox>(pb.authStore.model!.id)
        return json(r)
    } catch (e: any) {
        throw error(e.status || 500, e);
    }
}
