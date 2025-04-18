import GPX from "$lib/models/gpx/gpx";
import { haversineDistance } from "$lib/models/gpx/utils";
import { SummitLog } from "$lib/models/summit_log";
import type { Trail } from "$lib/models/trail";
import { pb } from "$lib/pocketbase";
import { fetchGPX, trails_create } from "$lib/stores/trail_store";
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
        let parseResult: { trail: Trail, gpx: GPX };
        try {
            parseResult = (await gpx2trail(gpxData, data.get("name") as string | undefined));
        } catch (e: any) {
            console.error(e)
            throw new ClientResponseError({ status: 400, response: { message: "Invalid file" } })
        }
        let trail = parseResult.trail;

        const ignoreDuplicates = data.get("ignoreDuplicates") === "true"
        if (!ignoreDuplicates) {
            let duplicate: Trail | null = null;
            try {
                duplicate = await findDuplicate(trail)
            } catch (e: any) {
                throw new ClientResponseError({ status: 500, response: { message: "Error checking for duplicates" } })
            }
            if (duplicate !== null) {
                throw new ClientResponseError({ status: 400, response: { message: `Duplicate trail`, id: duplicate.id, name: duplicate.name }, })
            }
        }

        // const log = new SummitLog(trail.date as string, {
        //     distance: trail.distance,
        //     elevation_gain: trail.elevation_gain,
        //     elevation_loss: trail.elevation_loss,
        //     duration: trail.duration ? trail.duration * 60 : undefined,
        // })
        // log.expand!.gpx_data = gpxData;
        // const fileName = (data.get("name") as string | null)?.length ? data.get("name") as string : "file"
        // log._gpx = new File([gpxFile], fileName);

        // trail.expand!.summit_logs?.push(log);


        try {
            trail = await trails_create(trail, [], gpxFile, event.fetch);
        } catch (e: any) {
            console.error(e)
            throw handleError(e)
        }
        return json(trail);

    } catch (e: any) {
        throw handleError(e)
    }
}

async function findDuplicate(t1: Trail) {
    const trails: Trail[] = await pb.collection('trails').getFullList({ requestKey: null });

    const distanceThreshold = 100;
    const elevationThreshhold = 50;
    const lengthThreshhold = 50;

    for (const t2 of trails) {
        const lengthDifference = Math.abs((t1.distance ?? 0) - (t2.distance ?? 0));
        const elevationGainDifference = Math.abs((t1.elevation_gain ?? 0) - (t2.elevation_gain ?? 0));
        const elevationLossDifference = Math.abs((t1.elevation_loss ?? 0) - (t2.elevation_loss ?? 0));
        const startpointDifference = haversineDistance(t1.lat ?? 0, t1.lon ?? 0, t2.lat ?? 0, t2.lon ?? 0)

        if (lengthDifference < lengthThreshhold && elevationGainDifference < elevationThreshhold && elevationLossDifference < elevationThreshhold && startpointDifference < distanceThreshold) {
            return t2
        }
    }

    return null
}
