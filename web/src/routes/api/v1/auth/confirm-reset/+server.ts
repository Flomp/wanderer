import { pb } from "$lib/pocketbase";
import { handleError } from "$lib/util/api_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";
import { z } from "zod";

export async function POST(event: RequestEvent) {
    try {
        const data = await event.request.json()
        const safeData = z.object({
            token: z.string(),
            password: z.string().min(8).max(72),
            passwordConfirm: z.string().min(8).max(72)
        }).refine(d => d.password === d.passwordConfirm).parse(data)
        const r = await pb.collection('users').confirmPasswordReset(safeData.token, safeData.password, safeData.passwordConfirm);
        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }

}
