import type { UserAnonymous } from "$lib/models/user";
import { Collection, handleError, show } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await show<UserAnonymous>(event, Collection.users_anonymous)

        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}