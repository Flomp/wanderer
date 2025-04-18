import { SettingsCreateSchema } from '$lib/models/api/settings_schema';
import type { Settings } from "$lib/models/settings";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";


export async function GET(event: RequestEvent) {
    try {
        const r = await show<Settings>(event, Collection.settings)
        return json(r)
    } catch (e: any) {
        throw handleError(e);
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<Settings>(event, SettingsCreateSchema, Collection.settings)
        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.settings)
        return json(r);
    } catch (e: any) {
        throw handleError(e);
    }
}