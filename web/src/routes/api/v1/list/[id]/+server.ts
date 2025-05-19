import { ListUpdateSchema } from "$lib/models/api/list_schema";
import type { List } from "$lib/models/list";
import { Collection, handleError, remove, show, update } from "$lib/util/api_util";
import { objectToFormData } from "$lib/util/file_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export async function GET(event: RequestEvent) {
    try {
        let l: List
        if (!event.url.searchParams.has("handle")) {
            l = await show<List>(event, Collection.lists)
        } else {
            const actor = await event.locals.pb.send(`/activitypub/actor?resource=acct:${event.url.searchParams.get("handle")}`, { method: "GET", fetch: event.fetch, });
            event.url.searchParams.delete("handle")

            if (actor.isLocal) {
                l = await show<List>(event, Collection.lists)
            } else {
                const origin = new URL(actor.iri).origin
                const url = `${origin}/api/v1/list/${event.params.id}`

                const response = await event.fetch(url + '?' + event.url.searchParams, { method: 'GET' })
                if (!response.ok) {
                    const errorResponse = await response.json()
                    console.error(errorResponse)
                    const cachedList = await event.locals.pb.collection("lists").getOne(`${event.params.id}`)
                    return json(cachedList)
                }
                l = await response.json()
                l.avatar = l.avatar ? `${origin}/api/v1/files/lists/${l.id}/${l.avatar}` : undefined

                l.author = actor.id!
                l.expand!.author = actor
                l.iri = url;

                l.expand?.trails?.forEach(t => {
                    t.gpx = t.gpx ? `${origin}/api/v1/files/trails/${t.id}/${t.gpx}` : undefined
                    t.photos = t.photos.map(p =>
                        `${origin}/api/v1/files/trails/${t.id}/${p}`
                    )
                    t.iri = t.iri || `${origin}/api/v1/trails/${t.id}`;
                })

                let dbList: List | undefined;
                try {
                    dbList = await event.locals.pb.collection("lists").getFirstListItem(`iri='${url}'`)
                } catch (e) {
                    if (!(e instanceof ClientResponseError) || e.status != 404) {
                        throw e
                    }
                }

                const formData = objectToFormData({ ...l, expand: undefined, trails: [] })
                if (l.avatar) {
                    const avatarURL = l.avatar
                    let response = await event.fetch(avatarURL, { method: "GET" })
                    const avatar = await response.blob()


                    formData.append("avatar", avatar)
                }
                if (dbList !== undefined) {
                    dbList = await event.locals.pb.collection("lists").update(dbList.id!, formData)
                } else {
                    dbList = await event.locals.pb.collection("lists").create(formData)
                }

                l.id = dbList!.id
            }

        }

        return json(l)
    } catch (e: any) {
        return handleError(e)
    }
}

export async function POST(event: RequestEvent) {
    try {
        const r = await update<List>(event, ListUpdateSchema, Collection.lists)
        return json(r);
    } catch (e: any) {
        return handleError(e)
    }
}

export async function DELETE(event: RequestEvent) {
    try {
        const r = await remove(event, Collection.lists)
        return json(r);
    } catch (e: any) {
        return handleError(e)
    }
}

