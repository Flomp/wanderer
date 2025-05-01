import { handleError } from "$lib/util/api_util";
import { json, RequestEvent } from '@sveltejs/kit';
import PocketBase from "PocketBase";

export async function GET(event: RequestEvent) {
    try {
        const r = await ((event.locals as any).pb as PocketBase).send("/.well-known/webfinger?" + event.url.searchParams, { method: "GET", fetch: event.fetch })
        return json(r)
    } catch (err) {
        handleError(err)
    }
}