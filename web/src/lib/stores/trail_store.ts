import type { SummitLog } from "$lib/models/summit_log";
import { Trail, type TrailFilter, type TrailFilterValues, type TrailSearchResult } from "$lib/models/trail";
import type { Waypoint } from "$lib/models/waypoint";
import { pb } from "$lib/pocketbase";
import { deepEqual } from "$lib/util/deep_util";
import { getFileURL } from "$lib/util/file_util";
import * as M from "maplibre-gl";
import type { Hits } from "meilisearch";
import { type ListResult, type RecordModel } from "pocketbase";
import { writable, type Writable } from "svelte/store";
import { summit_logs_create, summit_logs_delete, summit_logs_update } from "./summit_log_store";
import { waypoints_create, waypoints_delete, waypoints_update } from "./waypoint_store";
import { APIError } from "$lib/util/api_util";

let trails: Trail[] = []
export const trail: Writable<Trail> = writable(new Trail(""));

export const editTrail: Writable<Trail> = writable(new Trail(""));

export async function trails_index(perPage: number = 21, random: boolean = false, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/trail?' + new URLSearchParams({
        "perPage": perPage.toString(),
        expand: "category,waypoints,summit_logs",
        sort: random ? "@random" : "",
    }), {
        method: 'GET',
    })
    const response: ListResult<Trail> = await r.json()

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    trails = response.items
    return response.items;

}

export async function trails_recommend(size: number, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/trail/recommend?' + new URLSearchParams({
        "size": size.toString(),
    }), {
        method: 'GET',
    })
    const response: Trail[] = await r.json()

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return response;

}

export async function trails_search_filter(filter: TrailFilter, page: number = 1, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    let filterText: string = buildFilterText(filter, true);

    let r = await f("/api/v1/search/trails", {
        method: "POST",
        body: JSON.stringify({
            q: filter.q,
            options: {
                filter: filterText,
                sort: [`${filter.sort}:${filter.sortOrder == "+" ? "asc" : "desc"}`],
                hitsPerPage: 12,
                page: page
            }
        }),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const result: { page: number, totalPages: number, hits: Hits<TrailSearchResult> } = await r.json();

    if (result.hits.length == 0) {
        return { items: [], ...result };
    }

    const resultTrails: Trail[] = await searchResultToTrailList(result.hits)

    return { items: resultTrails, ...result };

}

export async function trails_search_bounding_box(northEast: M.LngLat, southWest: M.LngLat, filter?: TrailFilter, page: number = 1, loadGPX: boolean = true) {

    let filterText: string = "";

    if (filter) {
        filterText = buildFilterText(filter, false);
    }

    let r = await fetch("/api/v1/search/trails", {
        method: "POST",
        body: JSON.stringify({
            q: "",
            options: {
                filter: [
                    `_geoBoundingBox([${northEast.lat}, ${northEast.lng}], [${southWest.lat}, ${southWest.lng}])`,
                    filterText
                ],
                hitsPerPage: 500,
                page: page
            }
        }),
    });
    const result: { page: number, totalPages: number, hits: Hits<TrailSearchResult> } = await r.json();

    if (result.hits.length == 0) {
        trails = [];
        return { trails: [], ...result }
    }

    const resultTrails: Trail[] = await searchResultToTrailList(result.hits, loadGPX)

    trails = page > 1 ? trails.concat(resultTrails) : resultTrails

    return { trails, ...result };


}

export async function trails_show(id: string, loadGPX?: boolean, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f(`/api/v1/trail/${id}?` + new URLSearchParams({
        expand: "category,waypoints,summit_logs,trail_share_via_trail",
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const response = await r.json()

    if (loadGPX) {
        if (!response.expand) {
            response.expand = {}
        }
        const gpxData: string = await fetchGPX(response, f);
        if (!response.expand) {
            response.expand = {};
        }
        response.expand.gpx_data = gpxData;


        for (const log of response.expand.summit_logs ?? []) {
            const gpxData: string = await fetchGPX(log, f);

            if (!log.expand) {
                log.expand = {};
            }
            log.expand.gpx_data = gpxData;
        }
    }

    response.expand.waypoints = response.expand.waypoints || [];
    response.expand.summit_logs = response.expand.summit_logs?.sort((a: SummitLog, b: SummitLog) => Date.parse(a.date) - Date.parse(b.date)) || [];

    trail.set(response);

    return response as Trail;
}

export async function trails_create(trail: Trail, photos: File[], gpx: File | Blob | null, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {

    for (const waypoint of trail.expand?.waypoints ?? []) {
        const model = await waypoints_create({
            ...waypoint,
            marker: undefined,
        }, f);
        trail.waypoints.push(model.id!);
    }
    for (const summitLog of trail.expand?.summit_logs ?? []) {
        const model = await summit_logs_create(summitLog, f);
        trail.summit_logs.push(model.id!);
    }

    trail.author = pb.authStore.model!.id

    let r = await f('/api/v1/trail', {
        method: 'PUT',
        body: JSON.stringify({ ...trail, expand: undefined }),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }


    let model: Trail = await r.json();

    const formData = new FormData()
    if (gpx) {
        formData.append("gpx", gpx);
    }

    for (const photo of photos) {
        formData.append("photos", photo)
    }

    r = await f(`/api/v1/trail/${model.id!}/file`, {
        method: 'POST',
        body: formData,
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return await r.json();

}

export async function trails_update(oldTrail: Trail, newTrail: Trail, photos?: File[], gpx?: File | Blob | null) {

    const waypointUpdates = compareObjectArrays<Waypoint>(oldTrail.expand?.waypoints ?? [], newTrail.expand?.waypoints ?? []);

    for (const addedWaypoint of waypointUpdates.added) {
        const model = await waypoints_create({
            ...addedWaypoint,
            marker: undefined,
        },);
        newTrail.waypoints.push(model.id!);
    }

    for (const updatedWaypoint of waypointUpdates.updated) {
        const oldWaypoint = oldTrail.expand?.waypoints?.find(w => w.id == updatedWaypoint.id);
        const model = await waypoints_update(oldWaypoint!, {
            ...updatedWaypoint,
            marker: undefined,
        });
    }

    for (const deletedWaypoint of waypointUpdates.deleted) {
        const success = await waypoints_delete(deletedWaypoint);
    }

    const summitLogUpdates = compareObjectArrays<SummitLog>(oldTrail.expand?.summit_logs ?? [], newTrail.expand?.summit_logs ?? []);

    for (const summitLog of summitLogUpdates.added) {
        const model = await summit_logs_create(summitLog);
        newTrail.summit_logs.push(model.id!);
    }

    for (const updatedSummitLog of summitLogUpdates.updated) {
        const oldSummitLog = oldTrail.expand?.summit_logs?.find(w => w.id == updatedSummitLog.id);

        const model = await summit_logs_update(oldSummitLog!, updatedSummitLog);
    }

    for (const deletedSummitLog of summitLogUpdates.deleted) {
        const success = await summit_logs_delete(deletedSummitLog);
    }

    let r = await fetch(`/api/v1/trail/${newTrail.id}?` + new URLSearchParams({
        expand: "category,waypoints,summit_logs,trail_share_via_trail",
    }), {
        method: 'POST',
        body: JSON.stringify({ ...newTrail, expand: undefined }),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }


    let model: Trail = await r.json();

    const formData = new FormData()
    if (gpx) {
        formData.append("gpx", gpx);
    }

    if (photos) {
        for (const photo of photos) {
            formData.append("photos", photo)
        }
    }


    const deletedPhotos = oldTrail.photos.filter(oldPhoto => !newTrail.photos.find(newPhoto => newPhoto === oldPhoto));

    for (const deletedPhoto of deletedPhotos) {
        formData.append("photos-", deletedPhoto.replace(/^.*[\\/]/, ''));
    }

    r = await fetch(`/api/v1/trail/${newTrail.id!}/file`, {
        method: 'POST',
        body: formData,
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }



    trail.set(model);

    return model;
}


export async function trails_delete(trail: Trail) {
    if (trail.expand?.waypoints) {
        for (const waypoint of trail.expand.waypoints) {
            await waypoints_delete(waypoint);
        }
    }
    if (trail.expand?.summit_logs) {
        for (const summit_log of trail.expand.summit_logs) {
            await summit_logs_delete(summit_log);
        }
    }

    const r = await fetch('/api/v1/trail/' + trail.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }


    return await r.json();

}

export async function trails_get_filter_values(f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch): Promise<TrailFilterValues> {
    const r = await f('/api/v1/trail/filter', {
        method: 'GET',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return await r.json();

}

export async function trails_get_bounding_box(f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch): Promise<TrailFilterValues> {
    const r = await f('/api/v1/trail/bounding-box', {
        method: 'GET',
    })
    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return await r.json();

}

export async function trails_upload(file: File, ignoreDuplicates: boolean = false, onProgress?: (progress: number) => void) {
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        const fd = new FormData();
        fd.append("name", file.name);
        fd.append("file", file);
        fd.append("ignoreDuplicates", ignoreDuplicates ? "true" : "false")

        xhr.open("PUT", "/api/v1/trail/upload", true);

        xhr.upload.onprogress = function (event) {
            if (event.lengthComputable) {
                const percentComplete = (event.loaded / event.total) * 100;
                onProgress?.(percentComplete)
            }
        };

        xhr.onload = async () => {
            const responseText = xhr.responseText;
            const response = responseText ? JSON.parse(responseText) : null;

            if (xhr.status >= 200 && xhr.status < 300) {
                resolve(response);
            } else {
                reject(new APIError(xhr.status, response?.message || "Upload failed", response));
            }
        };

        xhr.onerror = () => {
            reject(new APIError(xhr.status, xhr.statusText));
        };

        xhr.send(fd);
    });
}

export async function fetchGPX(trail: { gpx?: string } & Record<string, any>, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    if (!trail.gpx) {
        return "";
    }
    const gpxUrl = getFileURL(trail, trail.gpx);
    const response: Response = await f(gpxUrl);
    const gpxData = await response.text();

    return gpxData
}

async function searchResultToTrailList(hits: Hits<TrailSearchResult>, loadGPX: boolean = false): Promise<Trail[]> {
    const trails: Trail[] = []
    for (const h of hits) {
        const t: Trail & RecordModel = {
            collectionId: "trails",
            collectionName: "trails",
            updated: new Date(h.created * 1000).toISOString(),
            author: h.author,
            name: h.name,
            photos: h.thumbnail ? [h.thumbnail] : [],
            public: h.public,
            summit_logs: [],
            waypoints: [],
            category: h.category,
            created: new Date(h.created * 1000).toISOString(),
            date: new Date(h.date * 1000).toISOString(),
            description: h.description,
            difficulty: h.difficulty,
            distance: h.distance,
            duration: h.duration,
            elevation_gain: h.elevation_gain,
            elevation_loss: h.elevation_loss,
            id: h.id,
            lat: h._geo.lat,
            lon: h._geo.lng,
            location: h.location,
            gpx: h.gpx,
            thumbnail: 0,
            expand: {
                author: {
                    collectionId: "users",
                    private: false,
                    id: h.author,
                    avatar: h.author_avatar,
                    username: h.author_name
                } as any,
                trail_share_via_trail: h.shares?.map(s => ({
                    permission: "view",
                    trail: h.id,
                    user: s,
                })),
            }
        }

        if (loadGPX) {
            const gpxData: string = await fetchGPX(t);
            t.expand!.gpx_data = gpxData;
        }
        
        trails.push(t)
    }
    return trails
}

function buildFilterText(filter: TrailFilter, includeGeo: boolean): string {
    let filterText: string = "";

    filterText += `distance >= ${Math.floor(filter.distanceMin)} AND elevation_gain >= ${Math.floor(filter.elevationGainMin)} AND elevation_loss >= ${Math.floor(filter.elevationLossMin)}`

    if (filter.distanceMax < filter.distanceLimit) {
        filterText += ` AND distance <= ${Math.ceil(filter.distanceMax)}`
    }

    if (filter.elevationGainMax < filter.elevationGainLimit) {
        filterText += ` AND elevation_gain <= ${Math.ceil(filter.elevationGainMax)}`
    }

    if (filter.elevationLossMax < filter.elevationLossLimit) {
        filterText += ` AND elevation_loss <= ${Math.ceil(filter.elevationLossMax)}`
    }

    if (filter.difficulty.length > 0) {
        filterText += ` AND difficulty IN [${filter.difficulty.join(",")}]`
    }

    if (filter.author?.length) {
        filterText += ` AND author = ${filter.author}`
    }

    if (filter.public !== undefined || filter.shared !== undefined) {
        filterText += " AND ("
        if (filter.public !== undefined) {
            filterText += `(public = ${filter.public}`

            if (!filter.author?.length || filter.author == pb.authStore.model?.id) {
                filterText += ` OR author = ${pb.authStore.model?.id}`
            }
            filterText += ")"
        }

        if (filter.shared !== undefined) {
            if (filter.shared === true) {
                filterText += ` OR shares = ${pb.authStore.model?.id}`
            } else {
                filterText += ` AND NOT shares = ${pb.authStore.model?.id}`

            }
        }
        filterText += ")"
    }

    if (filter.startDate) {
        filterText += ` AND date >= ${new Date(filter.startDate).getTime() / 1000}`
    }

    if (filter.endDate) {
        filterText += ` AND date <= ${new Date(filter.endDate).getTime() / 1000}`
    }

    if (filter.category.length > 0) {
        filterText += ` AND category IN [${filter.category.join(",")}]`;
    }
    if (filter.completed !== undefined) {
        filterText += ` AND completed = ${filter.completed}`;
    }

    if (filter.near.lat && filter.near.lon && includeGeo) {
        filterText += ` AND _geoRadius(${filter.near.lat}, ${filter.near.lon}, ${filter.near.radius})`
    }
    if (filter.near.lat && filter.near.lon && includeGeo) {
        filterText += ` AND _geoRadius(${filter.near.lat}, ${filter.near.lon}, ${filter.near.radius})`
    }

    return filterText
}

function compareObjectArrays<T extends { id?: string }>(oldArray: T[], newArray: T[]) {
    const newObjects = [];
    const updatedObjects = [];
    const unchangedObjects = [];
    for (const newObj of newArray) {
        const oldObj = oldArray.find(oldObj => oldObj.id === newObj.id)
        if (!oldObj) {
            newObjects.push(newObj);
        } else if (!deepEqual(newObj, oldObj)) {
            updatedObjects.push(newObj);
        } else {
            unchangedObjects.push(newObj);
        }
    }
    const deletedObjects = oldArray.filter(oldObj => !newArray.find(newObj => newObj.id === oldObj.id));

    return {
        added: newObjects,
        deleted: deletedObjects,
        updated: updatedObjects,
        unchanged: unchangedObjects,
    };
}
