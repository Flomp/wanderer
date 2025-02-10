import type { Category } from "$lib/models/category";
import { Collection, handleError, list } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Category>(event, Collection.categories);
        return json(r)
    } catch (e) {
        throw handleError(e)
    }

}