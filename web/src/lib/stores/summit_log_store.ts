import { SummitLog } from "$lib/models/summit_log";
import { pb } from "$lib/pocketbase";
import { ClientResponseError } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const summitLog: Writable<SummitLog> = writable(new SummitLog(new Date().toISOString().substring(0, 10)));

export async function summit_logs_create(summitLog: SummitLog) {
    summitLog.author = pb.authStore.model!.id

    let r = await fetch('/api/v1/summit-log', {
        method: 'PUT',
        body: JSON.stringify(summitLog),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    if (summitLog._gpx && summitLog._gpx instanceof File) {
        let model: SummitLog = await r.json();

        const formData = new FormData()

        formData.append("gpx", summitLog._gpx)

        r = await fetch(`/api/v1/summit-log/${model.id!}/file`, {
            method: 'POST',
            body: formData,
        })
    }

    if (r.ok) {
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }
}

export async function summit_logs_update(summitLog: SummitLog) {
    summitLog.author = pb.authStore.model!.id

    let r = await fetch('/api/v1/summit-log/' + summitLog.id, {
        method: 'POST',
        body: JSON.stringify(summitLog),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }


    if (summitLog._gpx) {
        const formData = new FormData()

        formData.append("gpx", summitLog._gpx);
        r = await fetch(`/api/v1/summit-log/${summitLog.id!}/file`, {
            method: 'POST',
            body: formData,
        })
    }

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