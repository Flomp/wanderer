import type { SummitLog } from "$lib/models/summit_log";
import { Trail, type TrailFilter } from "$lib/models/trail";
import type { Waypoint } from "$lib/models/waypoint";
import { pb } from "$lib/pocketbase";
import { getFileURL } from "$lib/util/file_util";
import { util } from "$lib/vendor/svelte-form-lib/util";
import { get, writable, type Writable } from "svelte/store";
import { summit_logs_create, summit_logs_delete, summit_logs_update } from "./summit_log_store";
import { waypoints_create, waypoints_delete, waypoints_update } from "./waypoint_store";
import { ms } from "$lib/meilisearch";
import type { LatLng } from "leaflet";

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

export async function trails_search_filter(filter: TrailFilter) {
    let filterText: string = `distance >= ${filter.distanceMin} AND distance <= ${filter.distanceMax} AND elevation_gain >= ${filter.eleavationGainMin} AND elevation_gain <= ${filter.elevationGainMax}`;

    if (filter.category.length > 0) {
        filterText += ` AND category IN [${filter.category.join(",")}]`;
    }
    if (filter.completed !== undefined) {
        filterText += ` AND completed = ${filter.completed}`;
    }
    if (filter.near.lat && filter.near.lon) {
        filterText += ` AND _geoRadius(${filter.near.lat}, ${filter.near.lon}, ${filter.near.radius})`
    }
    const indexResponse = await ms
        .index("trails")
        .search(filter.q, { filter: filterText });
    const trailIds = indexResponse.hits.map((h) => h.id);

    if (trailIds.length == 0) {
        trails.set([]);
        return [];
    }

    const dbResponse: Trail[] = (await pb.collection('trails').getList<Trail>(1, 5, {
        filter: trailIds.map((id) => `id="${id}"`).join('||'), expand: "category,waypoints,summit_logs",
        sort: `${filter.sortOrder}${filter.sort}`
    })).items

    for (const trail of dbResponse) {
        setFileURLs(trail);
    }

    trails.set(dbResponse);

    return dbResponse;
}

export async function trails_search_bounding_box(northEast: LatLng, southWest: LatLng) {
    const response = await ms.index("trails").search("", {
        filter: [
            `_geoBoundingBox([${northEast.lat}, ${northEast.lng}], [${southWest.lat}, ${southWest.lng}])`,
        ],
    });
    const trailIds = response.hits.map((h) => h.id);

    if (trailIds.length == 0) {
        trails.set([]);
        return compareObjectArrays<Trail>(get(trails), []);
    }

    const dbResponse: Trail[] = (
        await pb.collection("trails").getList<Trail>(1, 5, {
            filter: trailIds.map((id) => `id="${id}"`).join("||"),
            expand: "category,waypoints,summit_logs",
            sort: `+name`,
        })
    ).items;

    for (const trail of dbResponse) {
        setFileURLs(trail);
        const gpxData: string = await fetchGPX(trail);
        trail.expand.gpx_data = gpxData;
    }
    
    const comparison = compareObjectArrays<Trail>(get(trails), dbResponse)

    trails.set(dbResponse);

    return comparison;
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

    if (!pb.authStore.model) {
        throw new Error("Unauthenticated");
        ;
    }
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

    if (!formData.get("public")) {
        formData.set("public", 0);
    }

    formData.append("author", pb.authStore.model!.id);

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
        .update<Trail>(model.id!, { thumbnail: thumbnail }, { expand: "category" });

    index_trail(model);

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

    for (const updatedWaypoint of waypointUpdates.updated) {
        const model = await waypoints_update({
            ...updatedWaypoint,
            marker: undefined,
        });
    }

    for (const deletedWaypoint of waypointUpdates.deleted) {
        const success = await waypoints_delete(deletedWaypoint.id!);
    }

    const summitLogUpdates = compareObjectArrays<SummitLog>(oldTrail.expand.summit_logs ?? [], newTrail.expand.summit_logs ?? []);

    for (const summitLog of summitLogUpdates.added) {
        const model = await summit_logs_create(summitLog);
        formData.append("summit_logs+", model.id!);
    }

    for (const updatedSummitLog of summitLogUpdates.updated) {
        const model = await summit_logs_update(updatedSummitLog);
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

    if (!formData.get("public")) {
        formData.set("public", 0);
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
        .update<Trail>(model.id!, { thumbnail: thumbnail }, { expand: "category,waypoints,summit_logs" });

    trail.set(model);

    index_trail(model);

    return model;
}


export async function trails_delete(trail: Trail) {
    if (trail.expand.waypoints) {
        for (const waypoint of trail.expand.waypoints) {
            waypoints_delete(waypoint.id!);
        }
    }
    if (trail.expand.summit_logs) {
        for (const summit_log of trail.expand.summit_logs) {
            summit_logs_delete(summit_log.id!);
        }
    }

    ms.index('trails').deleteDocument(trail.id!)

    const success = await pb
        .collection("trails")
        .delete(trail.id!);

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
    const newObjects = [];
    const updatedObjects = [];
    for (const newObj of newArray) {
        const oldObj = oldArray.find(oldObj => oldObj.id === newObj.id)
        if (!oldObj) {
            newObjects.push(newObj);
        } else if (!util.deepEqual(newObj, oldObj)) {
            updatedObjects.push(newObj);
        }
    }
    const deletedObjects = oldArray.filter(oldObj => !newArray.find(newObj => newObj.id === oldObj.id));

    return {
        added: newObjects,
        deleted: deletedObjects,
        updated: updatedObjects
    };
}

function index_trail(trail: Trail) {
    ms.index('trails').addDocuments([
        {
            "id": trail.id,
            "name": trail.name,
            "description": trail.description,
            "location": trail.location,
            "distance": trail.distance,
            "elevation_gain": trail.elevation_gain,
            "duration": trail.duration,
            "category": trail.expand.category?.id,
            "completed": trail.expand.summit_logs?.length > 0,
            "created": trail.created,
            "_geo": {
                "lat": trail.lat,
                "lng": trail.lon,
            }
        }]);
}

