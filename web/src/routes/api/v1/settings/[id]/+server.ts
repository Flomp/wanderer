import type { Settings } from "$lib/models/settings";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";


export async function GET(event: RequestEvent) {
    try {
        const r = await pb.collection('settings').getFirstListItem<Settings>(`user="${event.params.id}"`)
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const r = await await pb.collection("settings").update<Settings>(event.params.id as string, data);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await pb.collection('settings').delete(event.params.id as string)
        return json({ 'acknowledged': r });
    } catch (e: any) {
        throw error(e.status, e)
    }
}