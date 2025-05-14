import { env } from '$env/dynamic/private';
import type { Actor } from '$lib/models/activitypub/actor';
import type { UserAnonymous } from "$lib/models/user";
import { actorFromRemote, splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { json, type RequestEvent } from '@sveltejs/kit';
import type { APActivity } from 'activitypub-types';
import { ClientResponseError } from 'pocketbase';

export async function POST(event: RequestEvent) {

    try {
        let fullUsername = event.params.username;
        if (!fullUsername) {
            return json("Bad request", { status: 400 });

        }

        fullUsername = fullUsername.replace(/^@/, "");

        const [username, domain] = splitUsername(fullUsername, env.ORIGIN)

        const activity: APActivity = await event.request.json()

        if (!activity.actor) {
            return json("Bad request", { status: 400 });
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

        const success = await event.locals.pb.send("/activitypub/activity/process", {
            method: "POST", fetch: event.fetch,
            headers: {
                'X-Forwarded-Path': event.url.pathname,
                signature: event.request.headers.get("signature")!,
                date: event.request.headers.get("date")!,
                digest: event.request.headers.get("digest")!
            },
            body: activity
        })

        if (success === false) {
            return json("Invalid header signature", { status: 400 });
        }

        const headers = new Headers()
        headers.append("Content-Type", "application/activity+json")

        return json("", { status: 200, headers });
    } catch (e) {
        return handleError(e)
    }


}

