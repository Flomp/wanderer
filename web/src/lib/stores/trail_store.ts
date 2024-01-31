import { pb } from "$lib/constants";
import { Trail } from "$lib/models/trail";
import { getFileURL } from "$lib/util/file_util";
import { writable, type Writable } from "svelte/store";
import { waypoints_create, waypoints_delete } from "./waypoint_store";
import { summit_logs_create, summit_logs_delete } from "./summit_log_store";
import type { Waypoint } from "$lib/models/waypoint";
import type { SummitLog } from "$lib/models/summit_log";

export const trails: Writable<Trail[]> = writable([])
export const trail: Writable<Trail> = writable(new Trail(""));

export const editTrail: Writable<Trail> = writable(new Trail(""));

export async function trails_index() {
    const response: Trail[] = (await pb.collection('trails').getList<Trail>(1, 5, { expand: "category,waypoints,summit_logs" })).items

    for (const trail of response) {
        setFileURLs(trail);
    }

    trails.set(response);

    return response;
}

export async function trails_show(id: string, loadGPX?: boolean) {
    const response: Trail = await pb.collection('trails').getOne<Trail>(id, { expand: "category,waypoints,summit_logs" })

    if (loadGPX) {
        const gpxData: string = await fetchGPX(response);
        response.expand.gpx_data = gpxData;
    }

    setFileURLs(response);

    response.expand.waypoints = response.expand.waypoints || [];
    response.expand.summit_logs = response.expand.summit_logs || [];

    response._photoFiles = [];

    trail.set(response);

    return response;
}

export async function trails_create(trail: Trail, formData: { [key: string]: any; } | FormData) {
    formData.set("category", trail.expand.category!.id);

    for (const file of trail._photoFiles) {
        formData.append("photos", file);
    }

    for (const waypoint of trail.expand.waypoints) {
        const model = await waypoints_create({
            ...waypoint,
            marker: undefined,
        });
        formData.append("waypoints", model.id!);
    }
    for (const summitLog of trail.expand.summit_logs) {
        const model = await summit_logs_create(summitLog);
        formData.append("summit_logs", model.id!);
    }
    let model = await pb
        .collection("trails")
        .create<Trail>(formData);

    const thumbnailIndex = trail.photos.findIndex(
        (p) => p == trail.thumbnail,
    );
    let thumbnail: string | undefined = "/imgs/thumbnail.jpg";
    if (thumbnailIndex >= 0) {
        thumbnail = model.photos.at(thumbnailIndex);
    }

    model = await pb
        .collection("trails")
        .update<Trail>(model.id!, { thumbnail: thumbnail });

    return model;
}

export async function trails_update(oldTrail: Trail, newTrail: Trail, formData: { [key: string]: any; } | FormData) {

    const waypointUpdates = compareObjectArrays<Waypoint>(oldTrail.expand.waypoints ?? [], newTrail.expand.waypoints ?? []);

    for (const addedWaypoint of waypointUpdates.added) {
        const model = await waypoints_create({
            ...addedWaypoint,
            marker: undefined,
        });
        formData.append("waypoints+", model.id!);
    }

    for (const deletedWaypoint of waypointUpdates.deleted) {
        const success = await waypoints_delete(deletedWaypoint.id!);
    }

    const summitLogUpdates = compareObjectArrays<SummitLog>(oldTrail.expand.summit_logs ?? [], newTrail.expand.summit_logs ?? []);

    for (const summitLog of summitLogUpdates.added) {
        const model = await summit_logs_create(summitLog);
        formData.append("summit_logs+", model.id!);
    }

    for (const deletedSummitLog of summitLogUpdates.deleted) {
        const success = await summit_logs_delete(deletedSummitLog.id!);
    }

    for (const file of newTrail._photoFiles) {
        formData.append("photos", file);
    }

    const deletedPhotos = oldTrail.photos.filter(oldPhoto => !newTrail.photos.find(newPhoto => newPhoto === oldPhoto));

    for (const deletedPhoto of deletedPhotos) {
        formData.append("photos-", deletedPhoto.replace(/^.*[\\/]/, ''));
    }

    if (formData.get("gpx").size == 0) {
        formData.delete("gpx");
    }

    const thumbnailIndex = newTrail.photos.findIndex(
        (p) => p == newTrail.thumbnail,
    );

    let model = await pb
        .collection("trails")
        .update<Trail>(newTrail.id!, formData);

    let thumbnail: string | undefined = oldTrail.thumbnail;
    if (thumbnailIndex >= 0) {
        thumbnail = model.photos.at(thumbnailIndex);
    }

    model = await pb
        .collection("trails")
        .update<Trail>(model.id!, { thumbnail: thumbnail });

    trail.set(model);

    // return model;
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

function setFileURLs(trail: Trail) {
    trail.thumbnail = getFileURL(trail, trail.thumbnail);
    for (let i = 0; i < trail.photos.length; i++) {
        const photo = trail.photos[i];
        trail.photos[i] = getFileURL(trail, photo)
    }
}

function compareObjectArrays<T extends { id?: string }>(oldArray: T[], newArray: T[]) {
    const newObjects = newArray.filter(newObj => !oldArray.find(oldObj => oldObj.id === newObj.id));
    const deletedObjects = oldArray.filter(oldObj => !newArray.find(newObj => newObj.id === oldObj.id));

    return {
        added: newObjects,
        deleted: deletedObjects
    };
}

