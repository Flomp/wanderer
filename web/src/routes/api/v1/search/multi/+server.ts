import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    const data = await event.request.json()

    try {
        const r = await event.locals.ms.multiSearch({
            queries: data.queries
        });
        return json(r);
    } catch (e: any) {
        throw error(e.httpStatus, e)
    }
}