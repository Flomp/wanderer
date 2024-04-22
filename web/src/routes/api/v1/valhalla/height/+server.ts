import { env } from '$env/dynamic/public';
import { error, json, type NumericRange, type RequestEvent } from "@sveltejs/kit";


export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    if (!env.PUBLIC_VALHALLA_URL) {
        return error(400, "PUBLIC_VALHALLA_URL not set")
    }
    try {
        const r = await event.fetch(env.PUBLIC_VALHALLA_URL + '/height', { method: "POST", body: JSON.stringify(data) });        
        const response = await r.json();
        if (!r.ok) {
            throw error(r.status as NumericRange<400,500>, response);
        }
        return json(response);
    } catch (e: any) {
        throw error(e.status || 500, e)
    }
}