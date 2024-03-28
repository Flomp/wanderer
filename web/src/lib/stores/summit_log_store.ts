import { SummitLog } from "$lib/models/summit_log";
import { ClientResponseError } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const summitLog: Writable<SummitLog> = writable(new SummitLog(new Date().toISOString().substring(0, 10)));

export async function summit_logs_create(summitLog: SummitLog) {
    const r = await fetch('/api/v1/summit-log', {
        method: 'PUT',
        body: JSON.stringify(summitLog),
    })

    if (r.ok) {
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }
}

export async function summit_logs_update(summitLog: SummitLog) {
    const r = await fetch('/api/v1/summit-log/' + summitLog.id, {
        method: 'POST',
        body: JSON.stringify(summitLog),
    })

    if (r.ok) {
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }
}

export async function summit_logs_delete(summitLog: SummitLog) {
    const r = await fetch('/api/v1/summit-log/' + summitLog.id, {
        method: 'DELETE',
    })

    if (r.ok) {
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }
}