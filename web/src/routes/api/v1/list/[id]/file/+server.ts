import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";
import type { List } from "postcss/lib/list";

export async function POST(event: RequestEvent) {
    const data = await event.request.formData()
    try {
        const r = await pb.collection("lists").update<List>(event.params.id as string, data,);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}