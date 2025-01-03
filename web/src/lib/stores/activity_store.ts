import type { Activity } from "$lib/models/activity";
import { APIError } from "$lib/util/api_util";
import { type ListResult } from "pocketbase";

let activities: Activity[] = [];

export async function activities_index(author: string, page: number = 1, perPage: number = 10, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/activity?' + new URLSearchParams({
        "perPage": perPage.toString(),
        page: page.toString(),
        filter: `author="${author}"`,
        sort: "-created"
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const fetchedActivities: ListResult<Activity> = await r.json();

    const result = page > 1 ? [...activities, ...fetchedActivities.items] : fetchedActivities.items

    activities = result;

    return { ...fetchedActivities, items: result };

}