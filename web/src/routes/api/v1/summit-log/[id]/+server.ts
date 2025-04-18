import { SummitLogUpdateSchema } from "$lib/models/api/summit_log_schema";
import type { SummitLog } from "$lib/models/summit_log";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";


export async function GET(event: RequestEvent) {
    try {
        const r = await show<SummitLog>(event, Collection.summit_logs)

        if (!r.expand) {
            r.expand = {} as any
        }
        r.expand!.author = await event.locals.pb.collection("users_anonymous").getOne(r.author!);

        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<SummitLog>(event, SummitLogUpdateSchema, Collection.summit_logs)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.summit_logs)
        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }
}