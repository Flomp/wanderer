import { APIError } from "$lib/util/api_util";
import { get } from "svelte/store";
import { currentUser } from "./user_store";

export async function actors_search(q: string, includeSelf: boolean = true) {
    const user = get(currentUser)

    let r = await fetch('/api/v1/activitypub/actor?' + new URLSearchParams({
        "filter": `username~"${q}"${includeSelf ? '' : `&&id!="${user?.actor}"`}`,
    }), {
        method: 'GET',
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const response = await r.json()
    return response.items;

}