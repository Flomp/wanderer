import type { List } from "$lib/models/list";
import { Collection, handleError, uploadUpdate } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    try {        
        const r = await uploadUpdate<List>(event, Collection.lists)
        return json(r);
    } catch (e) {
        return handleError(e)
    }
}
