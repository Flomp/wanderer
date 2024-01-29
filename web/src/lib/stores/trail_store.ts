import { pb } from "$lib/constants";
import { Trail } from "$lib/models/trail";
import { getFileURL } from "$lib/util/file_util";
import { writable, type Writable } from "svelte/store";

export const trails: Writable<Trail[]> = writable([])
export const trail: Writable<Trail> = writable(new Trail(""));

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

export async function trails_create(bodyParams?: { [key: string]: any; } | FormData) {
    const model = await pb
        .collection("trails")
        .create<Trail>(bodyParams);

    return model;
}

export async function trails_update(id: string, bodyParams?: { [key: string]: any; } | FormData) {
    const model = await pb
        .collection("trails")
        .update<Trail>(id, bodyParams);

    return model;
}


export async function trails_delete(id: string) {
    const success = await pb
        .collection("trails")
        .delete(id);

    return success;
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