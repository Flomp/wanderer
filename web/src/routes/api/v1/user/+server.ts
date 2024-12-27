import { env } from '$env/dynamic/public';
import { UserCreateSchema } from '$lib/models/api/user_schema';
import type { User } from '$lib/models/user';
import { Collection, create, handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';
import { ClientResponseError } from 'pocketbase';

export async function PUT(event: RequestEvent) {
    try {
        if (env.PUBLIC_DISABLE_SIGNUP === "true") {
            throw new ClientResponseError({ status: 401, response: { messgage: "Forbidden" } })
        }
        const r = await create<User>(event, UserCreateSchema, Collection.users)

        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }
}

