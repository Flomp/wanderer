import type { Activity } from "$lib/models/activity";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    const page = event.url.searchParams.get("page") ?? "0";
    const perPage = event.url.searchParams.get("per-page") ?? "10";
    const filter = event.url.searchParams.get("filter") ?? "";

    try {
        const r = await pb.collection('activities')
            .getList<Activity>(parseInt(page), parseInt(perPage), { filter: filter ?? "" })
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

