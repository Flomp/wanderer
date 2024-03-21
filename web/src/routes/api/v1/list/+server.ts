import type { List } from '$lib/models/list';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    const sort = event.url.searchParams.get('sort') ?? ""
    try {
        const r: List[] = await pb.collection('lists').getFullList<List>({
            expand: "trails,trails.waypoints,trails.category",
            sort: sort,
        })
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function PUT(event: RequestEvent) {
    const data = await event.request.json();

    try {
        const r = await pb.collection('lists').create<List>(data)
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}