import { invalidateAll } from "$app/navigation";
import { Settings } from "$lib/models/settings";
import { APIError } from "$lib/util/api_util";

export async function settings_show(userId: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/settings/' + userId, {
        method: 'GET',
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }
    const response = await r.json();

    return response
}

export async function settings_create(settings: Settings) {
    const r = await fetch('/api/v1/settings', {
        method: 'PUT',
        body: JSON.stringify(settings),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return await r.json();

}

export async function settings_update(settings: Settings) {
    const r = await fetch('/api/v1/settings/' + settings.id, {
        method: 'POST',
        body: JSON.stringify(settings),
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    await invalidateAll()
    return await r.json();

}

export async function settings_delete(settings: Settings) {
    const r = await fetch('/api/v1/settings/' + settings.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return await r.json();

}