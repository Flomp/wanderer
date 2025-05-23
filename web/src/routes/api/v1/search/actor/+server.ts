import type { Actor } from '$lib/models/activitypub/actor';
import { splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import { ClientResponseError, type ListResult } from "pocketbase"

export async function GET(event: RequestEvent) {
    try {

        if (!event.url.searchParams.has("q")) {
            throw new ClientResponseError({ status: 400, response: "Bad request" });

        }
        const q = event.url.searchParams.get("q")

        const [user, domain] = splitUsername(q!)

        const r = await event.fetch(`${domain ? `https://${domain}` : ''}/api/v1/activitypub/actor?` + new URLSearchParams({
            filter: `isLocal=true&&username~'${user}'&&user.settings_via_user.privacy.account != 'private'`,
            perPage: "3"
        }),)

        if (!r.ok) {
            const errorResponse = await r.json()
            throw new ClientResponseError({ status: r.status, response: errorResponse });
        }

        const response: ListResult<Actor> = await r.json()

        if(domain) {
            response.items.forEach(i => i.isLocal = false)
        }


        return json({ items: response.items })


    } catch (e) {
        if (e instanceof Error && e.message == "fetch failed") {
            return error(404, "Not found")
        }
        return handleError(e)
    }
}
