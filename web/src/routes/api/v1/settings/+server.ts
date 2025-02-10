import { SettingsCreateSchema } from '$lib/models/api/settings_schema';
import { Collection, create } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function PUT(event: RequestEvent) {
    try {
        const r = await create<Comment>(event, SettingsCreateSchema, Collection.settings)

        return json(r);
    } catch (e: any) {
        throw error(e.status, e)
    }
}
