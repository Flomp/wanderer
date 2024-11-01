import type { SummitLog } from "$lib/models/summit_log";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    const data = await event.request.formData()
    try {
        const r = await pb.collection("summit_logs").update<SummitLog>(event.params.id as string, data,);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}