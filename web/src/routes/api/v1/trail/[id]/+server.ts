import type { Trail } from "$lib/models/trail";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { error, json, type RequestEvent } from "@sveltejs/kit";
import { TrailUpdateSchema } from '$lib/models/api/trail_schema';
import type PocketBase from "pocketbase";

export async function GET(event: RequestEvent) {
    try {
        const r = await show<Trail>(event, Collection.trails)

        if (event.locals.pb.authStore.record) {
            if (!r.expand) {
                r.expand = {} as any
            }
            r.expand!.author = await event.locals.pb.collection("users_anonymous").getOne(r.author!);
        }

        // remove time from dates
        await enrichRecord(event.locals.pb, r);

        // sort waypoints by distance
        r.expand?.waypoints?.sort((a, b) => (a.distance_from_start ?? 0) - (b.distance_from_start ?? 0))
        return json(r)
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<Trail>(event, TrailUpdateSchema, Collection.trails)
        await enrichRecord(event.locals.pb, r)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.trails)
        return json(r);
    } catch (e: any) {
        throw handleError(e)
    }
}



async function enrichRecord(pb: PocketBase, r: Trail) {
    r.date = r.date?.substring(0, 10) ?? "";
    for (const log of r.expand?.summit_logs ?? []) {
        log.date = log.date.substring(0, 10);

        if (!log.expand) {
            log.expand = {} as any;
        }
        if (log.author) {
            log.expand!.author = await pb.collection("users_anonymous").getOne(log.author);
        }
    }
}