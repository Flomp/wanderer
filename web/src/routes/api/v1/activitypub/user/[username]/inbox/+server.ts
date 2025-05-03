import { env } from '$env/dynamic/private';
import type { Actor } from '$lib/models/activitypub/actor';
import type { UserAnonymous } from "$lib/models/user";
import { actorFromRemote, splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, type RequestEvent } from '@sveltejs/kit';
import type { APActivity } from 'activitypub-types';
import { ClientResponseError } from 'pocketbase';

export async function POST(event: RequestEvent) {

    try {
        let fullUsername = event.params.username;
        if (!fullUsername) {
            return error(400, "Bad request");
        }

        fullUsername = fullUsername.replace(/^@/, "");

        const [username, domain] = splitUsername(fullUsername, env.ORIGIN)

        const user: UserAnonymous = await event.locals.pb.collection("users_anonymous").getFirstListItem(`username='${username}'`)
        const object: Actor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`user='${user.id}'`)

        const activity: APActivity = await event.request.json()

        if (!activity.actor) {
            return error(400, "Bad request")
        }

        let actor: Actor;
        try {
            actor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`iri='${activity.actor.toString()}'`);
        } catch (e) {
            if (e instanceof ClientResponseError && e.status == 404) {
                // we have not seen this actor before:
                // discover its data via webfinger and save them to our db
                actor = (await actorFromRemote(domain, username, event.fetch)).actor
                await event.locals.pb.collection("activitypub_actors").create(actor);
            }
            throw e
        }

        const success = await event.locals.pb.send("/activitypub/signature/verify", {
            method: "POST", fetch: event.fetch,
            headers: {
                'X-Forwarded-Path': event.url.pathname,
                signature: event.request.headers.get("signature")!,
                date: event.request.headers.get("date")!,
                digest: event.request.headers.get("digest")!
            },
            body: {
                publicKey: actor.public_key
            }
        })

        if (!success) {
            return error(400, "Invalid header signature")
        }

        return;
    } catch (e) {
        throw handleError(e)
    }


}

