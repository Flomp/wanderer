import type { User } from "$lib/models/user";
import { Collection, upload } from "$lib/util/api_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function POST(event: RequestEvent) {
    try {
        const r = await upload<User>(event, Collection.users);
        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}