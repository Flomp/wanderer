import { Collection, handleError, show } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r = await event.locals.pb.collection(Collection.follow_counts)
            .getFirstListItem(`user = '${event.params.id}'`)
        return json(r)
    } catch (e: any) {
        throw handleError(e)
    }
}