import type { TimelineItem } from "$lib/models/timeline";
import { defaultTrailSearchAttributes, Trail, type TrailFilter, type TrailSearchResult } from "$lib/models/trail";
import { APIError } from "$lib/util/api_util";
import type { Hits } from "meilisearch";
import type { ListResult } from "pocketbase";
import { searchResultToTrailList } from "./trail_store";

let timeline: TimelineItem[] = []


export async function profile_show(username: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    let r = await f('/api/v1/profile/' + username, {
        method: 'GET',
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const response = await r.json()
    return response;

}

export async function profile_timeline_index(username: string, page: number, perPage: number = 1, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    let r = await f(`/api/v1/profile/${username}/timeline?` + new URLSearchParams({
        page: page.toString(),
        perPage: perPage.toString()
    }), {
        method: 'GET',
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const fetchedTimeline: ListResult<TimelineItem> = await r.json()

    const result = page > 1 ? [...timeline, ...fetchedTimeline.items] : fetchedTimeline.items

    timeline = result;

    return { ...fetchedTimeline, items: result };

}

export async function profile_trails_index(username: string, filter: TrailFilter, page: number = 1, perPage: number = 1, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    let r = await f(`/api/v1/profile/${username}/trails`, {
        method: "POST",
        body: JSON.stringify({
            q: filter.q,
            options: {
                attributesToRetrieve: defaultTrailSearchAttributes,
                sort: [`${filter.sort}:${filter.sortOrder == "+" ? "asc" : "desc"}`],
                hitsPerPage: 12,
                page: page
            }
        }),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const result: { page: number, totalPages: number, hits: Hits<TrailSearchResult> } = await r.json();

    if (result.hits.length == 0) {
        return { items: [], ...result };
    }

    const resultTrails: Trail[] = await searchResultToTrailList(result.hits)

    return { items: resultTrails, ...result };

}