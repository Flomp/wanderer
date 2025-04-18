import { handleError } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await event.locals.pb.send("/integration/komoot/login", {
            method: "GET",
        });
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}