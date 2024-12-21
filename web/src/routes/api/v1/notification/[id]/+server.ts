import type { Follow } from "$lib/models/follow";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";


export async function POST(event: RequestEvent) {
    const data = await event.request.json()

    try {
        const r = await pb.collection('notifications').update<Notification>(event.params.id as string, { ...data, seen: true })
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}