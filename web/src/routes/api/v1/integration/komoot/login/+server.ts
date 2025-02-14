import { pb } from "$lib/pocketbase";
import { handleError } from "$lib/util/api_util";
import { json } from "@sveltejs/kit";

export async function GET() {
    try {
        const r = await pb.send("/integration/komoot/login", {
            method: "GET",
        });
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}