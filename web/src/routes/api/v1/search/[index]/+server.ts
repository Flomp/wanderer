import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    const data = await event.request.json()

    try {
        const r = await event.locals.ms.index(event.params.index as string).search(data.q, data.options);       
        return json(r);
    } catch (e: any) {
        console.log(e);
        
        throw error(e.httpStatus, e)
    }
}