import { RecordListOptionsSchema } from "$lib/models/api/base_schema";
import { APIError, handleError } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    try {
        const data = await event.request.json();
        const searchParams = Object.fromEntries(event.url.searchParams);
        const safeSearchParams = RecordListOptionsSchema.parse(searchParams);


        const headers = new Headers()
        headers.append("Accept", "application/ld+json")

        const r = await event.fetch(data.outbox + '?' + new URLSearchParams({
            "perPage": safeSearchParams.perPage?.toString() ?? "10",
            page: safeSearchParams.page?.toString() ?? "1",
            filter: `type='Create'&&object.type='Trail'`,
        }), {
            headers: headers,
            method: 'GET',
        })
        if (!r.ok) {
            const response = await r.json();
            throw new APIError(r.status, response.message, response.detail)

        }
        const response = await r.json();
        return json(response)
    } catch (e) {
        return handleError(e)
    }
}

