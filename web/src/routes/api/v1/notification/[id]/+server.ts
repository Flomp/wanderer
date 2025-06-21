import { NotificationUpdateSchema } from "$lib/models/api/notification_schema";
import { Collection, handleError, update } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";


export async function POST(event: RequestEvent) {
    try {
        const r = await update<Comment>(event, NotificationUpdateSchema, Collection.notifications)
        return json(r);
    } catch (e: any) {
        return handleError(e)
    }
}