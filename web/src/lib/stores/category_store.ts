import { pb } from "$lib/pocketbase";
import type { Category } from "$lib/models/category";
import { writable, type Writable } from "svelte/store";
import { ClientResponseError } from "pocketbase";

export const categories: Writable<Category[]> = writable([])

export async function categories_index(f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/category', {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const response = await r.json();

    categories.set(response);

    return response;
}