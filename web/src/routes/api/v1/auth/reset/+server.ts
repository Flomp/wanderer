import { handleError } from "$lib/util/api_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";
import { z } from "zod";

export async function POST(event: RequestEvent) {
    try {
        const data = await event.request.json()
        const safeData = z.object({
            email: z.string().email()
        }).parse(data)

        const r = await event.locals.pb.collection('users').requestPasswordReset(safeData.email);
        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }

}
