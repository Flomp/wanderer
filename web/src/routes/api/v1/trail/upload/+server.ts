import type { Trail } from "$lib/models/trail";
import { trails_create } from "$lib/stores/trail_store";
import { fromFIT, fromKML, fromTCX, gpx2trail, isFITFile } from "$lib/util/gpx_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export async function PUT(event: RequestEvent) {
    try {
        const data = await event.request.blob();
        const fileBuffer = await data.arrayBuffer();
        const fileContent = await data.text();
        let gpxData = ""
        let gpxFile: Blob;
        if (isFITFile(fileBuffer)) {           
            gpxData = await fromFIT(fileBuffer);            
            gpxFile = new Blob([gpxData], {
                type: "application/gpx+xml",
            });
        }
        else if (fileContent.includes("http://www.opengis.net/kml")) {
            gpxData = fromKML(fileContent);
            gpxFile = new Blob([gpxData], {
                type: "application/gpx+xml",
            });
        } else if (fileContent.includes("TrainingCenterDatabase")) {
            gpxData = fromTCX(fileContent);
            gpxFile = new Blob([gpxData], {
                type: "application/gpx+xml",
            });
        } else {
            gpxData = fileContent;
            gpxFile = data
        }
        if (!gpxData.length) {
            throw new ClientResponseError({ status: 400, response: { message: "Empty file" } })
        }
        let trail: Trail;
        try {
            trail = (await gpx2trail(gpxData)).trail;
        } catch (e: any) {
            throw new ClientResponseError({ status: 400, response: { message: "Invalid file" } })
        }

        try {
            trail = await trails_create(trail, [], gpxFile, event.fetch);
        } catch (e: any) {
            console.log(e);

            throw e
        }
        return json(trail);

    } catch (e: any) {
        console.log(e);

        throw error(e.status, e)
    }
}