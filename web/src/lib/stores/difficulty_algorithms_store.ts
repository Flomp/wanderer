import type { DifficultyAlgorithm } from "$lib/models/difficulty_algorithms";
import { APIError } from "$lib/util/api_util";
import { type ListResult } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const algorithms: Writable<DifficultyAlgorithm[]> = writable([])

export async function algorithms_index(f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/algorithms', {
        method: 'GET',
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const response: ListResult<DifficultyAlgorithm> = await r.json();

    algorithms.set(response.items);

    return response.items as DifficultyAlgorithm[];
}
