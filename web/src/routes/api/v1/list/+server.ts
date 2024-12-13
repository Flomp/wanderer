import type { List } from '$lib/models/list';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    const page = event.url.searchParams.get("page") ?? "0";
    const perPage = event.url.searchParams.get("per-page") ?? "10";
    const sort = event.url.searchParams.get('sort') ?? ""
    const filter = event.url.searchParams.get("filter") ?? "";

    const expand = "trails,trails.waypoints,trails.category,list_share_via_list";

    try {
        let r;
        if (parseInt(perPage) < 0) {
            r = {
                items: await pb.collection('lists')
                    .getFullList<List>({ expand: expand, sort: sort, filter: filter })
            }
        } else {
            r = await pb.collection('lists')
                .getList<List>(parseInt(page), parseInt(perPage), { expand: expand, sort: sort ?? "", filter: filter ?? "" })
        }

        for (const t of r.items) {
            if (!t.author || !pb.authStore.model) {
                continue;
            }
            if (!t.expand) {
                t.expand = {}
            }
            t.expand!.author = await pb.collection("users_anonymous").getOne(t.author);
        }

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