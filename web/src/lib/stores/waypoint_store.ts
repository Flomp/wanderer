import { Waypoint } from "$lib/models/waypoint";
import { writable, type Writable } from "svelte/store";

export const waypoint: Writable<Waypoint> = writable(new Waypoint(0, 0));
