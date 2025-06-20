import { env } from '$env/dynamic/private';

import type { Actor } from '$lib/models/activitypub/actor';
import type { Follow } from '$lib/models/follow';
import { splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import type { APOrderedCollectionPage, APRoot } from 'activitypub-types';
import type { ListResult } from 'pocketbase';


export async function GET(event: RequestEvent) {

    try {
        const page = event.url.searchParams.get("page") ?? "1"

        const intPage = parseInt(page);

        let fullUsername = event.params.handle;
        if (!fullUsername) {
            return json({ message: "Bad request" }, { status: 400 });
        }

        fullUsername = fullUsername.replace(/^@/, "");

        const [username, domain] = splitUsername(fullUsername, env.ORIGIN)

        const actor: Actor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`username:lower='${username?.toLowerCase()}'&&isLocal=true`)

        const followers: ListResult<Follow> = await event.locals.pb.collection("follows").getList(intPage, 10, { sort: "-created", filter: `followee='${actor.id}'&&status='accepted'`, expand: "follower" })

        const id = actor.iri;

        const hasNextPage = intPage * 10 < followers.totalItems;

        const followersPage: APRoot<APOrderedCollectionPage> = {
            "@context": [
                "https://www.w3.org/ns/activitystreams",
            ],
            type: "OrderedCollectionPage",
            first: id + "/followers?page=1",
            ...(intPage > 1 ? { prev: `${id}/followers?page=${intPage - 1}` } : {}),
            ...(hasNextPage ? { next: `${id}/outbox?page=${intPage + 1}` } : {}),
            partOf: id + "/followers",
            totalItems: followers.totalItems,
            orderedItems: followers.items.map<string>(f => f.expand!.follower.iri)
        }

        const headers = new Headers()
        headers.append("Content-Type", "application/activity+json")

        return json(followersPage, { headers });
    } catch (e) {
        return handleError(e)
    }


}

