import { env } from '$env/dynamic/public';
import { UserCreateSchema } from '$lib/models/api/user_schema';
import type { User } from '$lib/models/user';
import { pb } from '$lib/pocketbase';
import { Collection, create, handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';
import { ClientResponseError } from 'pocketbase';

export async function PUT(event: RequestEvent) {
    if (env.PUBLIC_DISABLE_SIGNUP === "true") {
        throw new ClientResponseError({ status: 401, response: { messgage: "Forbidden" } })
    }

    try {

        const data = await event.request.json();
        const safeData = UserCreateSchema.parse(data);

        const r = await pb.collection(Collection.users).create<User>(safeData)

        await pb.collection('users').requestVerification(safeData.email);

        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }
}

