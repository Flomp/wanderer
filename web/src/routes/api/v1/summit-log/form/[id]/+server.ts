import type { SummitLog } from "$lib/models/summit_log";
import { Collection, handleError, uploadUpdate } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    try {
        const r = await uploadUpdate<SummitLog>(event, Collection.summit_logs)
        r.date = r.date?.substring(0, 10) ?? "";
        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}
