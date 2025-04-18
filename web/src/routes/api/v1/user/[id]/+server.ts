import { RecordIdSchema } from '$lib/models/api/base_schema';
import { UserUpdateSchema } from '$lib/models/api/user_schema';
import type { User } from '$lib/models/user';
import { Collection, handleError, remove, show } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    try {
        const r = await show<User>(event, Collection.users)

        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}

export async function POST(event: RequestEvent) {
    const data = await event.request.json()
    try {
        const params = event.params
        const safeParams = RecordIdSchema.parse(params);

        const safeData = UserUpdateSchema.parse(data);

        if (safeData.email && safeData.email != event.locals.pb.authStore.record!.email) {
            const r = await event.locals.pb.collection('users').requestEmailChange(safeData.email);
            event.locals.pb.authStore.record!.email = safeData.email;
        }
        const r = await event.locals.pb.collection('users').update<User>(safeParams.id, safeData)
        if (safeData.password) {
            const r = await event.locals.pb.collection('users').authWithPassword(safeData.email ?? safeData.username!, safeData.password);
            return json(r.record)
        } else {
            return json(r);
        }
    } catch (e: any) {
        throw handleError(e);
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.users)
        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }
}
