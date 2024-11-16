import { SummitLog, type SummitLogFilter } from "$lib/models/summit_log";
import { pb } from "$lib/pocketbase";
import { ClientResponseError } from "pocketbase";
import { writable, type Writable } from "svelte/store";
import { fetchGPX } from "./trail_store";

export const summitLog: Writable<SummitLog> = writable(new SummitLog(new Date().toISOString().substring(0, 10)));
export const summitLogs: Writable<SummitLog[]> = writable([]);

export async function summit_logs_index(author: string, filter?: SummitLogFilter, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {

    let filterText = `author='${author}'`
    filterText += filter ? "&&" + buildFilterText(filter) : "";

    const r = await f('/api/v1/summit-log?' + new URLSearchParams({
        filter: filterText,
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const fetchedSummitLogs: SummitLog[] = await r.json();

    for (const log of fetchedSummitLogs) {
        if (!log.gpx) {
            continue
        }
        const gpxData: string = await fetchGPX(log as any, f);

        if (!log.expand) {
            log.expand = {};
        }
        log.expand.gpx_data = gpxData;
    }

    summitLogs.set(fetchedSummitLogs);
    
    return fetchedSummitLogs;
}

export async function summit_logs_create(summitLog: SummitLog, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    summitLog.author = pb.authStore.model!.id

    let r = await f('/api/v1/summit-log', {
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

        r = await f(`/api/v1/summit-log/${model.id!}/file`, {
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