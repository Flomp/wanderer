import { handleError } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import { z } from "zod";

export async function POST(event: RequestEvent) {
    try {
        const data = await event.request.json();
        const safeData = z.object({
            email: z.string().email().optional(),
            username: z.string().optional(),
            password: z.string().min(8).max(72)
        }).refine(d => d.email !== undefined || d.username !== undefined).parse(data);


        const r = await event.locals.pb.collection('users').authWithPassword(safeData.email ?? safeData.username!, data.password);
        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }

}
