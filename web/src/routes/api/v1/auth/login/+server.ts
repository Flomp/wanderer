import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const r = await pb.collection('users').authWithPassword(data.email ?? data.username!, data.password);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e);
    }

}
