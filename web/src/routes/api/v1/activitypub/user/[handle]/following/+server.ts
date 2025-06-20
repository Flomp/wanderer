import { env } from '$env/dynamic/private';
import { env as publicEnv } from '$env/dynamic/public';

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

        const followers: ListResult<Follow> = await event.locals.pb.collection("follows").getList(intPage, 10, { sort: "-created", filter: `follower='${actor.id}'&&status='accepted'`, expand: "followee" })

        const id = actor.iri;

        const hasNextPage = intPage * 10 < followers.totalItems;

        const following: APRoot<APOrderedCollectionPage> = {
            "@context": [
                "https://www.w3.org/ns/activitystreams",
            ],
            type: "OrderedCollectionPage",
            first: id + "/following?page=1",
            ...(intPage > 1 ? { prev: `${id}/following?page=${intPage - 1}` } : {}),
            ...(hasNextPage ? { next: `${id}/following?page=${intPage + 1}` } : {}),
            partOf: id + "/following",
            totalItems: followers.totalItems,
            orderedItems: followers.items.map<string>(f => f.expand!.followee.iri)
        }

        const headers = new Headers()
        headers.append("Content-Type", "application/activity+json")

        return json(following, { headers });
    } catch (e) {
        return handleError(e)
    }


}

