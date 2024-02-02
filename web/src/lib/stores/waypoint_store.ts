import { pb } from "$lib/pocketbase";
import { Waypoint } from "$lib/models/waypoint";
import { writable, type Writable } from "svelte/store";

export const waypoint: Writable<Waypoint> = writable(new Waypoint(0, 0));

export async function waypoints_create(bodyParams?: { [key: string]: any; } | FormData) {

    const model = await pb
        .collection("waypoints")
        .create<Waypoint>(bodyParams);

    return model;
}

export async function waypoints_update(updatedWaypoint: Waypoint) {
    const model = await pb
        .collection("waypoints")
        .update(updatedWaypoint.id!, updatedWaypoint);

    return model;
}

export async function waypoints_delete(id: string) {
    const success = await pb
        .collection("waypoints")
        .delete(id);

    return success;
}