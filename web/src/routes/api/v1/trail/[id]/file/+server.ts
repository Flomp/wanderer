import type { Trail } from "$lib/models/trail";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    const data = await event.request.formData()
    try {
        const r = await pb.collection("trails").update<Trail>(event.params.id as string, data,);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}