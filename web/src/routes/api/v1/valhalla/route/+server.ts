import { env } from '$env/dynamic/public';
import { error, json, type NumericRange, type RequestEvent } from "@sveltejs/kit";


export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    if (!env.PUBLIC_VALHALLA_URL) {
        return json({ message: "PUBLIC_VALHALLA_URL not set" }, { status: 400 })
    }
    try {
        const r = await event.fetch(env.PUBLIC_VALHALLA_URL + '/route', { method: "POST", body: JSON.stringify(data) });
        const response = await r.json();
        if (!r.ok) {
            return json({ message: response }, { status: r.status })

        }
        return json(response);
    } catch (e: any) {
        return json({ message: e }, { status: 500 })
    }
}