import { SummitLog } from "$lib/models/summit_log";
import type { Trail } from "$lib/models/trail";
import { trails_create } from "$lib/stores/trail_store";
import { handleError } from "$lib/util/api_util";
import { fromFile, gpx2trail } from "$lib/util/gpx_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export async function PUT(event: RequestEvent) {
    try {
        const data = await event.request.formData();

        const { gpxData, gpxFile } = await fromFile(data.get("file") as Blob)

        if (!gpxData.length) {
            throw new ClientResponseError({ status: 400, response: { message: "Empty file" } })
        }
        let trail: Trail;
        try {
            trail = (await gpx2trail(gpxData, data.get("name") as string | undefined)).trail;

            const log = new SummitLog(trail.date as string, {
                distance: trail.distance,
                elevation_gain: trail.elevation_gain,
                elevation_loss: trail.elevation_loss,
                duration: trail.duration ? trail.duration * 60 : undefined,
            })
            log.expand!.gpx_data = gpxData;
            log._gpx = new File([gpxFile], (data.get("name") as string | null) ?? "file");

            trail.expand!.summit_logs.push(log);
        } catch (e: any) {
            console.error(e)
            throw new ClientResponseError({ status: 400, response: { message: "Invalid file" } })
        }

        try {
            trail = await trails_create(trail, [], gpxFile, event.fetch);
        } catch (e: any) {
            throw handleError(e)
        }
        return json(trail);

    } catch (e: any) {
        throw handleError(e)
    }
}