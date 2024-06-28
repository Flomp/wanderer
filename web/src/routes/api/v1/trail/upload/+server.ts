import type { Trail } from "$lib/models/trail";
import { trails_create } from "$lib/stores/trail_store";
import { gpx2trail } from "$lib/util/gpx_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export async function PUT(event: RequestEvent) {
    try {
        const data = await event.request.blob();
        const gpxString = await data.text();
        if (!gpxString.length) {
            throw new ClientResponseError({ status: 400, response: { message: "Empty file" } })
        }
        let trail: Trail;
        try {
            trail = (await gpx2trail(gpxString)).trail;
        } catch (e: any) {
            throw new ClientResponseError({ status: 400, response: { message: "Invalid GPX file" } })
        }

        try {
            trail = await trails_create(trail, [], data, event.fetch);
        } catch (e: any) {
            console.log(e);

            throw e
        }
        return json(trail);

    } catch (e: any) {
        throw error(e.status, e)
    }
}