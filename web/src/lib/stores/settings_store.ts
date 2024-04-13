import { invalidateAll } from "$app/navigation";
import { Settings } from "$lib/models/settings";
import { ClientResponseError } from "pocketbase";

export async function settings_show(userId: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/settings/' + userId, {
        method: 'GET',
    })

    const response = await r.json();
    if (r.ok) {
        return response;
    } else {
        throw new ClientResponseError(response)
    }
}

export async function settings_create(settings: Settings) {
    const r = await fetch('/api/v1/settings', {
        method: 'PUT',
        body: JSON.stringify(settings),
    })

    if (r.ok) {
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }
}

export async function settings_update(settings: Settings) {
    const r = await fetch('/api/v1/settings/' + settings.id, {
        method: 'POST',
        body: JSON.stringify(settings),
    })

    if (r.ok) {
        await invalidateAll()
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }
}

export async function settings_delete(settings: Settings) {
    const r = await fetch('/api/v1/settings/' + settings.id, {
        method: 'DELETE',
    })

    if (r.ok) {
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }
}