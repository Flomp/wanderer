import type { ListShare } from "$lib/models/list_share";
import { APIError } from "$lib/util/api_util";
import { type ListResult } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const shares: Writable<ListShare[]> = writable([])


export async function list_share_index(list: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/list-share?' + new URLSearchParams({
        filter: `list='${list}'`,
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const response: ListResult<ListShare> = await r.json();

    shares.set(response.items);

    return response;

}

export async function list_share_create(share: ListShare) {
    let r = await fetch('/api/v1/list-share', {
        method: 'PUT',
        body: JSON.stringify(share),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }
}

export async function list_share_update(share: ListShare) {
    let r = await fetch('/api/v1/list-share/' + share.id, {
        method: 'POST',
        body: JSON.stringify(share),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }
}

export async function list_share_delete(share: ListShare) {
    const r = await fetch('/api/v1/list-share/' + share.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }
}