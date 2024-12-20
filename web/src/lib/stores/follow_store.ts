import type { Follow } from "$lib/models/follow";
import { ClientResponseError, type ListResult } from "pocketbase";

let follows: Follow[] = [];

export async function follows_index(data: { follower?: string, followee?: string }, page: number = 1, perPage: number = 10, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/follow?' + new URLSearchParams({
        filter: data.follower ? `follower='${data.follower}'` : data.followee ? `followee='${data.followee}'` : '',
        page: page.toString(),
        "per-page": perPage.toString()
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const fetchedFollows: ListResult<Follow> = await r.json();

    const result = page > 1 ? [...follows, ...fetchedFollows.items] : fetchedFollows.items

    follows = result;

    return { ...fetchedFollows, items: result };
}

export async function follows_a_b(a: string, b: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/follow?' + new URLSearchParams({
        filter: `follower='${a}'&&followee='${b}'`,
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const response: ListResult<Follow> = await r.json();

    return response.items.at(0);
}

export async function follows_counts(id: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/follow/count/' + id, {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const response: { followers: number, following: number } = await r.json();

    return response;
}

export async function follows_create(follow: Follow) {
    let r = await fetch('/api/v1/follow', {
        method: 'PUT',
        body: JSON.stringify(follow),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}

export async function follows_update(follow: Follow) {
    let r = await fetch('/api/v1/follow/' + follow.id, {
        method: 'POST',
        body: JSON.stringify(follow),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}

export async function follows_delete(follow: Follow) {
    const r = await fetch('/api/v1/follow/' + follow.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}