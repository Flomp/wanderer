import { TagUpdateSchema } from "$lib/models/api/tag_schema";
import type { Tag } from "$lib/models/tag";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await show<Tag>(event, Collection.tags)
        return json(r)
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<Tag>(event, TagUpdateSchema, Collection.tags)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.tags)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}
