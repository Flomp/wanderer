import { pb } from "$lib/constants";
import type { Trail } from "$lib/models/trail";
import { getFileURL } from "$lib/util/file_util";
import { writable, type Writable } from "svelte/store";

export const trails: Writable<Trail[]> = writable()
export const trail: Writable<Trail> = writable();

export async function trails_index() {
    const response: Trail[] = (await pb.collection('trails').getList<Trail>(1, 5, { expand: "category,waypoints,summit_logs" })).items

    trails.set(response);
}

export async function trails_show(id: string, loadGPX?: boolean) {
    const response: Trail = await pb.collection('trails').getOne<Trail>(id, { expand: "category,waypoints,summit_logs" })

    if (loadGPX) {
        const gpxData: string = await fetchGPX(response);
        response.gpx = gpxData;
    }

    trail.set(response);
}

async function fetchGPX(trail: Trail) {
    if (!trail.gpx) {
        return "";
    }
    const gpxUrl = getFileURL(trail, trail.gpx);
    const response: Response = await fetch(gpxUrl);
    const gpxData = await response.text();

    return gpxData
}