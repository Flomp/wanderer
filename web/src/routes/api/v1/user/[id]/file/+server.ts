import { pb } from "$lib/pocketbase";
import type { User } from "$lib/stores/user_store";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    const data = await event.request.formData()
    try {
        const r = await pb.collection("users").update<User>(event.params.id as string, data,);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}