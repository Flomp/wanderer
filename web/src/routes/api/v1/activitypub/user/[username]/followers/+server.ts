import { env } from '$env/dynamic/private';
import type { Actor } from '$lib/models/activitypub/actor';
import type { Follow } from '$lib/models/follow';
import type { UserAnonymous } from "$lib/models/user";
import { splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import type { APActivity, APOrderedCollectionPage, APRoot } from 'activitypub-types';
import type { ListResult } from 'pocketbase';


export async function GET(event: RequestEvent) {

    try {
        const page = event.url.searchParams.get("page") ?? "1"

        const intPage = parseInt(page);

        let fullUsername = event.params.username;
        if (!fullUsername) {
            return error(400, "Bad request");
        }

        fullUsername = fullUsername.replace(/^@/, "");

        const [username, domain] = splitUsername(fullUsername, env.ORIGIN)

        const actor: Actor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`username='${username}'&&isLocal=true`)

        const followers: ListResult<Follow> = await event.locals.pb.collection("follows").getList(intPage, 10, { sort: "-created", filter: `followee='${actor.id}'`, expand: "follower" })

        const id = actor.iri;

        const hasNextPage = intPage * 10 < followers.totalItems;

        const outbox: APRoot<APOrderedCollectionPage> = {
            "@context": [
                "https://www.w3.org/ns/activitystreams",
            ],
            type: "OrderedCollectionPage",
            first: id + "/outbox?page=1",
            ...(intPage > 1 ? { prev: `${id}/outbox?page=${intPage - 1}` } : {}),
            ...(hasNextPage ? { next: `${id}/outbox?page=${intPage + 1}` } : {}),
            partOf: id + "/outbox",
            totalItems: followers.totalItems,
            orderedItems: followers.items.map<string>(f => f.expand!.follower.iri)
        }

        const headers = new Headers()
        headers.append("Content-Type", "application/activity+json")

        return json(outbox, { headers });
    } catch (e) {
        return handleError(e)
    }


}

