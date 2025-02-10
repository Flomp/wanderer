import { IntegrationCreateSchema } from "$lib/models/api/integration_schema";
import type { Integration } from "$lib/models/integration";
import { Collection, create, handleError, list } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await list<Integration>(event, Collection.integrations);

        return json(r)
    } catch (e) {
        throw handleError(e)
    }
}

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<Comment>(event, IntegrationCreateSchema, Collection.integrations)
        return json(r);
    } catch (e) {
        throw handleError(e)
    }
}