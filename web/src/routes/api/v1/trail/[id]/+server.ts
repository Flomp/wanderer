import { RecordOptionsSchema } from '$lib/models/api/base_schema';
import { TrailUpdateSchema } from '$lib/models/api/trail_schema';
import type { Trail } from "$lib/models/trail";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { objectToFormData } from "$lib/util/file_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import type PocketBase from "pocketbase";
import { ClientResponseError } from "pocketbase";

export async function GET(event: RequestEvent) {
    try {
        let t: Trail;
        if (!event.url.searchParams.has("handle")) {
            t = await show<Trail>(event, Collection.trails)
            return json(t);
        } else {
            const {actor, error} = await event.locals.pb.send(`/activitypub/actor?resource=acct:${event.url.searchParams.get("handle")}`, { method: "GET", fetch: event.fetch, });
            event.url.searchParams.delete("handle")

            const safeSearchParams = RecordOptionsSchema.parse(Object.fromEntries(event.url.searchParams));

            const origin = new URL(actor.iri).origin
            const url = `${origin}/api/v1/trail/${event.params.id}`

            let dbTrail: Trail | undefined;
            try {
                dbTrail = await event.locals.pb.collection("trails").getFirstListItem(`iri='${url}'||id='${event.params.id}'`, {
                    ...safeSearchParams
                })
            } catch (e) {
                if (!(e instanceof ClientResponseError) || e.status != 404) {
                    throw e
                }
            }

            if (actor.isLocal && dbTrail) {
                t = dbTrail
            } else {
                const response = await event.fetch((dbTrail?.iri ?? url)  + '?' + event.url.searchParams, { method: 'GET' })
                if (!response.ok) {
                    const errorResponse = await response.json()
                    console.error(errorResponse)
                    return json(dbTrail)
                }
                t = await response.json()
                t.gpx = `${origin}/api/v1/files/trails/${t.id}/${t.gpx}`
                t.photos = t.photos.map(p =>
                    `${origin}/api/v1/files/trails/${t.id}/${p}`
                )
                t.expand?.summit_logs_via_trail?.forEach(l => {
                    if (l.gpx) {
                        l.gpx = `${origin}/api/v1/files/summit_logs/${l.id}/${l.gpx}`
                    }
                    l.photos = l.photos.map(p =>
                        `${origin}/api/v1/files/summit_logs/${l.id}/${p}`
                    )

                    if (l.expand?.author) {
                        l.expand.author.isLocal = false
                    }
                })
                t.expand?.waypoints?.forEach(w => {

                    w.photos = w.photos.map(p =>
                        `${origin}/api/v1/files/waypoints/${w.id}/${p}`
                    )
                })
                t.author = actor.id!
                t.expand!.author = actor
                t.iri = dbTrail?.iri ?? url;

                const formData = objectToFormData({ ...t, id: dbTrail?.id, gpx: undefined, photos: [], waypoints: [], tags: [], category: undefined })
                if (t.photos.length) {
                    const photoURL = t.photos[t.thumbnail ?? 0]
                    let response = await event.fetch(photoURL, { method: "GET" })
                    const photo = await response.blob()
                    formData.append("photos", photo)

                    const gpxURL = t.gpx
                    response = await event.fetch(gpxURL, { method: "GET" })
                    const gpx = await response.blob()
                    formData.append("gpx", gpx)
                }
                if (dbTrail !== undefined) {
                    dbTrail = await event.locals.pb.collection("trails").update(dbTrail.id!, formData)
                } else {
                    dbTrail = await event.locals.pb.collection("trails").create(formData)
                }

                t.id = dbTrail!.id
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