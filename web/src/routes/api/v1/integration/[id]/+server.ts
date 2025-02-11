import { IntegrationUpdateSchema } from "$lib/models/api/integration_schema";
import type { Integration } from "$lib/models/integration";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await show<Integration>(event, Collection.integrations)
        return json(r)
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<Integration>(event, IntegrationUpdateSchema, Collection.integrations)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.integrations)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}
