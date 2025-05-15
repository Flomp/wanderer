import type { Trail } from "$lib/models/trail";
import { Collection, upload } from "$lib/util/api_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    try {
        const r = await upload<Trail>(event, Collection.trails);

        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}