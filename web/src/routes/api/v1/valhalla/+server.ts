import { error, json, type RequestEvent } from "@sveltejs/kit";
import { env } from '$env/dynamic/public'


export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    if(!env.PUBLIC_VALHALLA_URL) {
        return error(400, "PUBLIC_VALHALLA_URL not set")
    }
    try {
        const r = await event.fetch(env.PUBLIC_VALHALLA_URL + '/route', data);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}