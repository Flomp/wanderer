import type { FeedItem } from "$lib/models/feed";
import { Collection, handleError, list } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await list<FeedItem>(event, Collection.feed);
        return json(r)
    } catch (e) {
        return handleError(e)
    }

}