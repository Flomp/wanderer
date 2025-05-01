import { env } from "$env/dynamic/private";
import type { WebfingerResponse } from '$lib/models/activitypub/webfinger_response';
import type { Profile } from '$lib/models/profile';
import { splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import { ClientResponseError } from 'pocketbase';
import { type APActor, type APImage, type APOrderedCollection } from 'activitypub-types';
import type { Actor } from "$lib/models/activitypub/actor";

export async function GET(event: RequestEvent) {
    const fullUsername = event.params.username;
    if (!fullUsername) {
        return error(400, { message: "Bad request" })
    }
    const [username, domain] = splitUsername(fullUsername, env.ORIGIN)

    try {
        const actor: Actor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`domain='${domain}'&&username='${username}'`)

        const profile: Profile = {
            username: actor.username,
            acct: fullUsername,
            createdAt: actor.published ?? "",
            bio: actor.summary ?? "",
            uri: actor.IRI,
            followers: actor.followerCount ?? 0,
            following: actor.followingCount ?? 0,
            icon: actor.icon ?? "",
        }

        return json(profile)
    } catch (e) {
        // actor does not exist yet
    }

    try {
        const { actor, followers, following } = await fetchRemoteActor(event, domain, username);
        const profile: Profile = {
            username: actor.name ?? actor.preferredUsername ?? "",
            acct: `${username}@${domain}`,
            createdAt: actor.published?.toString() ?? "",
            bio: actor.summary ?? "",
            uri: actor.id ?? "",
            followers: followers?.totalItems ?? 0,
            following: following?.totalItems ?? 0,
            icon: (actor.icon as APImage)?.url?.toString() ?? "",
        };

        const cachedActor: Actor = {
            domain: domain,
            follower: actor.followers?.toString() ?? "",
            inbox: actor.inbox?.toString(),
            IRI: actor.id!.toString(),
            isLocal: false,
            last_fetched: new Date().toISOString(),
            username: actor.preferredUsername ?? "",
            followerCount: followers?.totalItems,
            followingCount: following?.totalItems,
            following: actor.following?.toString(),
            summary: actor.summary,
            outbox: actor.outbox.toString(),
            published: actor.published?.toString(),
            icon: (actor.icon as APImage)?.url?.toString() ?? "",
            public_key: actor.publicKey.publicKeyPem
        }

        await event.locals.pb.collection("activitypub_actors").create(cachedActor);
        return json(profile)
    } catch (e) {
        throw handleError(e)
    }
}

async function fetchRemoteActor(event: RequestEvent, domain?: string, username?: string) {

    let followers: APOrderedCollection | undefined;
    let following: APOrderedCollection | undefined;

    const webfingerURI = `https://${domain}/.well-known/webfinger?resource=acct:${username}@${domain}`;
    const webfingerRequest = await event.fetch(webfingerURI, { method: "GET" });

    if (!webfingerRequest.ok) {
        const errorResponse = await webfingerRequest.json();
        throw new ClientResponseError({ status: 500, response: errorResponse });
    }

    const webfinger: WebfingerResponse = await webfingerRequest.json();

    const actorURI = webfinger.links.find(l => l.rel == "self");
    if (!actorURI) {
        throw new ClientResponseError({ status: 500, response: { message: "webfinger response contains no actor URI" } });

    }
    const headers = new Headers();
    headers.append("Accept", "application/ld+json");
    const actorRequest = await event.fetch(actorURI.href, { method: "GET", headers: headers });

    if (!actorRequest.ok) {
        const errorResponse = await actorRequest.json();
        throw new ClientResponseError({ status: 500, response: errorResponse });
    }
    const apActor: APActor & { publicKey: { id: string, owner: string, publicKeyPem: string } } = await actorRequest.json();

    if (apActor.followers) {
        const followersRequest = await event.fetch(apActor.followers.toString(), { method: "GET", headers: headers });
        if (!followersRequest.ok) {
            const errorResponse = await webfingerRequest.json();
            throw new ClientResponseError({ status: 500, response: errorResponse });
        }
        followers = await followersRequest.json();
    }

    if (apActor.following) {
        const followingRequest = await event.fetch(apActor.following.toString(), { method: "GET", headers: headers });
        if (!followingRequest.ok) {
            const errorResponse = await webfingerRequest.json();
            throw new ClientResponseError({ status: 500, response: errorResponse });
        }
        following = await followingRequest.json();
    }



    return { actor: apActor, followers, following };
}
