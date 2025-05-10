import type { Activity } from "$lib/models/activitypub/activity";
import { APIError } from "$lib/util/api_util";
import type { APOrderedCollectionPage } from "activitypub-types";

let activities: Activity[] = [];

export async function activities_index(outbox: string, page: number = 1, perPage: number = 10, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {

    const r = await f("/api/v1/activity?" + new URLSearchParams({
        "perPage": perPage.toString(),
        page: page.toString(),
        filter: `type='Create'&&object.type='Trail'`,
    }), {
        body: JSON.stringify({outbox}),
        method: 'POST',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const fetchedActivities: APOrderedCollectionPage = await r.json();

    const result = page > 1 ? [...activities, ...(fetchedActivities.orderedItems as Activity[]) ?? []] : fetchedActivities.orderedItems as Activity[]

    activities = result;

    return { ...fetchedActivities, items: result };

}