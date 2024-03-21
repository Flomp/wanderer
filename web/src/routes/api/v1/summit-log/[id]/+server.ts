import type { SummitLog } from "$lib/models/summit_log";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";


export async function GET(event: RequestEvent) {
    try {
        const r = await pb.collection('summit_logs')
            .getOne<SummitLog>(event.params.id as string)
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }
}

export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const r = await await pb.collection("summit_logs").update<SummitLog>(event.params.id as string, data);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await pb.collection('summit_logs').delete(event.params.id as string)
        return json({ 'acknowledged': r });
    } catch (e: any) {
        throw error(e.status, e)
    }
}