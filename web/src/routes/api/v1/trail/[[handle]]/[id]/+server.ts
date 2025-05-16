import { TrailUpdateSchema } from '$lib/models/api/trail_schema';
import type { Trail } from "$lib/models/trail";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { objectToFormData } from "$lib/util/file_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import type PocketBase from "pocketbase";
import { ClientResponseError } from "pocketbase";

export async function GET(event: RequestEvent) {
    try {

        let t: Trail
        if (!event.params.handle) {
            t = await show<Trail>(event, Collection.trails)
        } else {
            const actor = await event.locals.pb.send(`/activitypub/actor?resource=acct:${event.params.handle}`, { method: "GET", fetch: event.fetch, });

            if (actor.isLocal) {
                t = await show<Trail>(event, Collection.trails)
            } else {
                const origin = new URL(actor.iri).origin
                const url = `${origin}/api/v1/trail/${event.params.id}`

                const response = await event.fetch(url + '?' + event.url.searchParams, { method: 'GET' })
                if (!response.ok) {
                    const errorResponse = await response.json()
                    throw new ClientResponseError({ status: response.status, response: errorResponse });
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
                t.author = actor.id!
                t.expand!.author = actor
                t.remote_url = url;

                let dbTrail: Trail | undefined;
                try {
                    dbTrail = await event.locals.pb.collection("trails").getFirstListItem(`remote_url='${url}'`)
                } catch (e) {
                    if (!(e instanceof ClientResponseError) || e.status != 404) {
                        throw e
                    }
                }

                const formData = objectToFormData({ ...t, gpx: undefined, photos: [], waypoints: [], tags: [], category: undefined })
                if (t.photos.length) {
                    const photoURL = t.photos[t.thumbnail ?? 0]
                    const response = await event.fetch(photoURL, { method: "GET" })
                    const photo = await response.blob()
                    formData.append("photos", photo)
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