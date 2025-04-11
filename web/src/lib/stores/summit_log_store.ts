import { SummitLog, type SummitLogFilter } from "$lib/models/summit_log";
import { pb } from "$lib/pocketbase";
import { type ListResult } from "pocketbase";
import { writable, type Writable } from "svelte/store";
import { fetchGPX } from "./trail_store";
import { APIError } from "$lib/util/api_util";

export const summitLog: Writable<SummitLog> = writable(new SummitLog(new Date().toISOString().substring(0, 10)));
export const summitLogs: Writable<SummitLog[]> = writable([]);

export async function summit_logs_index(author: string, filter?: SummitLogFilter, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {

    let filterText = `author='${author}'`
    filterText += filter ? "&&" + buildFilterText(filter) : "";

    const r = await f('/api/v1/summit-log?' + new URLSearchParams({
        filter: filterText,
        perPage: "-1",
        expand: "trails_via_summit_logs.category",
        sort: "+date",
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const fetchedSummitLogs: ListResult<SummitLog> = await r.json();

    // for (const log of fetchedSummitLogs.items) {
    //     if (!log.gpx) {
    //         continue
    //     }
    //     const gpxData: string = await fetchGPX(log as any, f);

    //     if (!log.expand) {
    //         log.expand = {};
    //     }
    //     log.expand.gpx_data = gpxData;
    // }

    summitLogs.set(fetchedSummitLogs.items);

    return fetchedSummitLogs;
}

export async function summit_logs_create(summitLog: SummitLog, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    summitLog.author = pb.authStore.record!.id

    let r = await f('/api/v1/summit-log', {
        method: 'PUT',
        body: JSON.stringify(summitLog),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    let model: SummitLog = await r.json();

    if (summitLog._gpx && summitLog._gpx instanceof File) {

        const formData = new FormData()

        formData.append("gpx", summitLog._gpx)

        r = await f(`/api/v1/summit-log/${model.id!}/file`, {
            method: 'POST',
            body: formData,
        })

        if (!r.ok) {
            const response = await r.json();
            throw new APIError(r.status, response.message, response.detail)
        }
    }

    if (summitLog._photos && summitLog._photos.length) {

        const formData = new FormData()

        for (const photo of summitLog._photos) {
            formData.append("photos", photo)
        }

        r = await fetch(`/api/v1/summit-log/${model.id!}/file`, {
            method: 'POST',
            body: formData,
        })

        if (!r.ok) {
            const response = await r.json();
            throw new APIError(r.status, response.message, response.detail)
        }
    }


    return model;
}

export async function summit_logs_update(oldSummitLog: SummitLog, newSummitLog: SummitLog) {
    newSummitLog.author = pb.authStore.record!.id

    let r = await fetch('/api/v1/summit-log/' + newSummitLog.id, {
        method: 'POST',
        body: JSON.stringify(newSummitLog),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const formData = new FormData()

    for (const photo of newSummitLog._photos ?? []) {
        formData.append("photos", photo)
    }

    const deletedPhotos = oldSummitLog.photos.filter(oldPhoto => !newSummitLog.photos.find(newPhoto => newPhoto === oldPhoto));

    for (const deletedPhoto of deletedPhotos) {
        formData.append("photos-", deletedPhoto.replace(/^.*[\\/]/, ''));
    }

    if (newSummitLog._gpx) {
        formData.append("gpx", newSummitLog._gpx);
    }

    r = await fetch(`/api/v1/summit-log/${newSummitLog.id!}/file`, {
        method: 'POST',
        body: formData,
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return await r.json();
}

export async function summit_logs_delete(summitLog: SummitLog) {
    const r = await fetch('/api/v1/summit-log/' + summitLog.id, {
        method: 'DELETE',
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return await r.json();

}

function buildFilterText(filter: SummitLogFilter,): string {
    let filterText: string = "";

    if (filter.category.length > 0) {
        filterText += `trails_via_summit_logs.category!=null&&'${filter.category.join(",")}'~trails_via_summit_logs.category`;
    }

    if (filter.startDate) {
        filterText += `${filter.category.length ? '&&' : ''}date>='${filter.startDate}'`
    }

    if (filter.endDate) {
        filterText += `${filter.category.length || filter.startDate ? '&&' : ''}date<='${filter.endDate}'`
    }

    return filterText;

}