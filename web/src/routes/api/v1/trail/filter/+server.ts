import { type TrailFilterValues } from '$lib/models/trail';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    if (!pb.authStore.model) {
        return json({
            min_distance: 0,
            max_distance: 20000,
            min_elevation_gain: 0,
            max_elevation_gain: 4000
        });
    }
    try {
        const r = await pb.collection('trails_filter').getOne<TrailFilterValues>(pb.authStore.model!.id)
        return json(r)
    } catch (e: any) {
        throw error(e.status || 500, e);
    }
}
