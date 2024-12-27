import type { SummitLog } from "$lib/models/summit_log";
import { Collection, upload } from "$lib/util/api_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    try {
        const r = await upload<SummitLog>(event, Collection.summit_logs);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}