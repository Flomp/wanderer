import type { Trail } from '$lib/models/trail';
import { pb } from '$lib/pocketbase';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    const page = event.url.searchParams.get("page") ?? "0";
    const perPage = event.url.searchParams.get("per-page") ?? "21";
    const expand = event.url.searchParams.get("expand") ?? ""
    const sort = event.url.searchParams.get("sort") ?? ""
    const filter = event.url.searchParams.get("filter") ?? "";

    try {
        let r;
        if (parseInt(perPage) < 0) {
            r = {
                items: await pb.collection('trails')
                    .getFullList<Trail>({ expand: expand, sort: sort, filter: filter })
            }
        } else {
            r = await pb.collection('trails')
                .getList<Trail>(parseInt(page), parseInt(perPage), { expand: expand ?? "", sort: sort ?? "", filter: filter ?? "" })
        }

        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function PUT(event: RequestEvent) {
    const data = await event.request.json();
    try {
        const r = await pb.collection('trails').create<Trail>(data)

        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}