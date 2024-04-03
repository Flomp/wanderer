import { type TrailFilterValues } from '$lib/models/trail';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await pb.collection('trails_filter').getOne<TrailFilterValues>(pb.authStore.model!.id)
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}
