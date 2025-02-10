import { Collection, handleError, show } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await show(event, Collection.follow_counts)
        return json(r)
    } catch (e: any) {
        throw handleError(e)
    }
}