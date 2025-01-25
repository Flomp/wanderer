

import { env } from "$env/dynamic/private";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    const data = await event.request.json()

    try {
        const r = await event.fetch(`${env.NOMINATIM_URL}/search?q=${data.q}&format=geocodejson&limit=${data.limit}`)
        return json(r);
    } catch (e: any) {
        console.log(e);

        throw error(e.httpStatus, e)
    }
}