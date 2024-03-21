import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export async function POST(event: RequestEvent) {
    try {
        let data;
        try {
            data = await event.request.json()
        } catch (e) {
            throw new ClientResponseError({ status: 400, response: { message: "Invalid json" } })
        }

        const r = await pb.collection('users').authWithPassword(data.email ?? data.username!, data.password);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e);
    }

}
