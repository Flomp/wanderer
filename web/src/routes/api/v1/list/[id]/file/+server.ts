import { Collection, handleError, upload } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import type { List } from "postcss/lib/list";

export async function POST(event: RequestEvent) {
    try {
        const r = await upload<List>(event, Collection.lists);
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}