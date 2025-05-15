import type { Trail } from "$lib/models/trail";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";
import { TrailUpdateSchema } from '$lib/models/api/trail_schema';
import type PocketBase from "pocketbase";
import { actorFromDb, splitUsername } from "$lib/util/activitypub_util";
import { ClientResponseError } from "pocketbase";

export async function GET(event: RequestEvent) {
    try {

        let t: Trail
        if (!event.params.handle) {
            t = await show<Trail>(event, Collection.trails)
        } else {
            const actor = await actorFromDb(event.locals.pb, event.params.handle, event.fetch);

            if (actor.isLocal) {
                t = await show<Trail>(event, Collection.trails)
            } else {
                const origin = new URL(actor.iri).origin
                const url = `${origin}/api/v1/trail/${event.params.id}?` + event.url.searchParams
                const response = await event.fetch(url, { method: 'GET' })
                if (!response.ok) {
                    const errorResponse = await response.json()
                    throw new ClientResponseError({ status: 500, response: errorResponse });
                }
                t = await response.json()
                t.gpx = `${origin}/api/v1/files/trails/${t.id}/${t.gpx}`
                t.photos = t.photos.map(p =>
                    `${origin}/api/v1/files/trails/${t.id}/${p}`
                )
                t.author = actor.id!
                
            }

        }

        // remove time from dates
        await enrichRecord(event.locals.pb, t);

        // sort waypoints by distance
        t.expand?.waypoints?.sort((a, b) => (a.distance_from_start ?? 0) - (b.distance_from_start ?? 0))
        return json(t)
    } catch (e: any) {
        return handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<Trail>(event, TrailUpdateSchema, Collection.trails)
        await enrichRecord(event.locals.pb, r)
        return json(r);
    } catch (e: any) {
        return handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.trails)
        return json(r);
    } catch (e: any) {
        return handleError(e)
    }
}



async function enrichRecord(pb: PocketBase, r: Trail) {
    r.date = r.date?.substring(0, 10) ?? "";
    for (const log of r.expand?.summit_logs_via_trail ?? []) {
        log.date = log.date.substring(0, 10);
    }
}