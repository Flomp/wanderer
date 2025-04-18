import { ListUpdateSchema } from "$lib/models/api/list_schema";
import type { List } from "$lib/models/list";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await show<List>(event, Collection.lists)
        if (!r.expand) {
            r.expand = {} as any
        }
        r.expand!.author = await event.locals.pb.collection('users_anonymous').getOne(r.author!)

        return json(r)
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<List>(event, ListUpdateSchema, Collection.lists)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.lists)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}

