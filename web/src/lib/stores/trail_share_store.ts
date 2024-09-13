import type { TrailShare } from "$lib/models/trail_share";
import { ClientResponseError } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const shares: Writable<TrailShare[]> = writable([])


export async function trail_share_index(data: {trail?: string, user?: string}, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/trail-share?' + new URLSearchParams({
        filter: data.trail ? `trail='${data.trail}'` : data.user ? `user='${data.user}'` : '',
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const response: TrailShare[] = await r.json();

    shares.set(response);

    return response;

}

export async function trail_share_create(share: TrailShare) {
    let r = await fetch('/api/v1/trail-share', {
        method: 'PUT',
        body: JSON.stringify(share),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}

export async function trail_share_update(share: TrailShare) {
    let r = await fetch('/api/v1/trail-share/' + share.id, {
        method: 'POST',
        body: JSON.stringify(share),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}

export async function trail_share_delete(share: TrailShare) {
    const r = await fetch('/api/v1/trail-share/' + share.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}