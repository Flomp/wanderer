import { ListShareUpdateSchema } from "$lib/models/api/list_share_schema";
import type { ListShare } from "$lib/models/list_share";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await show<ListShare>(event, Collection.list_share)
        return json(r)
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<ListShare>(event, ListShareUpdateSchema, Collection.list_share)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.list_share)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}
