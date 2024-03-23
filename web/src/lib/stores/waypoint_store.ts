import { Waypoint } from "$lib/models/waypoint";
import { pb } from "$lib/pocketbase";
import { ClientResponseError } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const waypoint: Writable<Waypoint> = writable(new Waypoint(0, 0));

export async function waypoints_create(waypoint: Waypoint, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {

    waypoint.author = pb.authStore.model!.id

    let r = await f('/api/v1/waypoint', {
        method: 'PUT',
        body: JSON.stringify(waypoint),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    if (waypoint._photos && waypoint._photos.length) {
        let model: Waypoint = await r.json();

        const formData = new FormData()

        for (const photo of waypoint._photos) {
            formData.append("photos", photo)
        }

        r = await fetch(`/api/v1/waypoint/${model.id!}/file`, {
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

export async function waypoints_update(oldWaypoint: Waypoint, newWaypoint: Waypoint) {
    newWaypoint.author = pb.authStore.model!.id

    let r = await fetch('/api/v1/waypoint/' + newWaypoint.id, {
        method: 'POST',
        body: JSON.stringify(newWaypoint),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const formData = new FormData()

    for (const photo of newWaypoint._photos ?? []) {
        formData.append("photos", photo)
    }

    const deletedPhotos = oldWaypoint.photos.filter(oldPhoto => !newWaypoint.photos.find(newPhoto => newPhoto === oldPhoto));

    for (const deletedPhoto of deletedPhotos) {
        formData.append("photos-", deletedPhoto.replace(/^.*[\\/]/, ''));
    }

    r = await fetch(`/api/v1/waypoint/${newWaypoint.id!}/file`, {
        method: 'POST',
        body: formData,
    })


    if (r.ok) {
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }
}

export async function waypoints_delete(waypoint: Waypoint) {
    const r = await fetch('/api/v1/waypoint/' + waypoint.id, {
        method: 'DELETE',
    })

    if (r.ok) {
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }
}