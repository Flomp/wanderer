import { Collection, handleError, list, show } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import type { DifficultyAlgorithm } from '$lib/models/difficulty_algorithms';

export async function GET(event: RequestEvent) {
    try {
        const r = await list<DifficultyAlgorithm>(event, Collection.difficulty_algorithms)
        return json(r)
    } catch (e) {
        return handleError(e);
    }
}
