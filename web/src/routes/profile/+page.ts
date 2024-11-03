import { summit_logs_index, summitLogs } from "$lib/stores/summit_log_store";
import { fetchGPX } from "$lib/stores/trail_store";
import { type ServerLoad } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load: ServerLoad = async ({ params, locals, fetch }) => {
    const logs = await summit_logs_index(fetch);

    for (const log of logs) {
        if (!log.gpx) {
            continue
        }
        const gpxData: string = await fetchGPX(log as any, fetch);

        if (!log.expand) {
            log.expand = {};
        }
        log.expand.gpx_data = gpxData;
    }

    console.log(logs);
    
};