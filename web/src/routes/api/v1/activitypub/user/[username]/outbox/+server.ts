import { env } from '$env/dynamic/private';
import type { Activity } from '$lib/models/activitypub/activity';
import type { Actor } from '$lib/models/activitypub/actor';
import type { UserAnonymous } from "$lib/models/user";
import { splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import type { APActivity, APOrderedCollection, APOrderedCollectionPage } from 'activitypub-types';
import type { ListResult } from 'pocketbase';


export async function GET(event: RequestEvent) {

    try {
        const page = event.url.searchParams.get("page") ?? "1"

        let fullUsername = event.params.username;
        if (!fullUsername) {
            return error(400, "Bad request");
        }

        fullUsername = fullUsername.replace(/^@/, "");

        const [username, domain] = splitUsername(fullUsername, env.ORIGIN)

        const user: UserAnonymous = await event.locals.pb.collection("users_anonymous").getFirstListItem(`username='${username}'`)
        const actor: Actor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`user='${user.id}'`)
        const activities: ListResult<Activity> = await event.locals.pb.collection("activitypub_activities").getList(parseInt(page), 10, { sort: "-created", filter: `actor='${actor.id}'` })

        const id = actor.iri;

        const orderedCollection: APOrderedCollection = {
            type: "OrderedCollection",
            id: id + '/outbox',
        }

        const outbox: APOrderedCollectionPage = {
            type: "OrderedCollectionPage",
            first: id + "/outbox?page=1",
            totalItems: activities.totalItems,
            orderedItems: activities.items
        }


        return json(outbox);
    } catch (e) {
        throw handleError(e)
    }


}

