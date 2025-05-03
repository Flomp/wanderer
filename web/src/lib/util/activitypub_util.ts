import type { Actor } from '$lib/models/activitypub/actor';
import type { WebfingerResponse } from '$lib/models/activitypub/webfinger_response';
import { type RequestEvent } from '@sveltejs/kit';
import { type APActor, type APImage, type APOrderedCollection } from 'activitypub-types';
import { ClientResponseError } from 'pocketbase';


export function splitUsername(username: string, localDomain: string) {
    const cleaned = username.replace(/^@/, "").trim();

    if (!cleaned.includes("@")) {
        return [cleaned, localDomain];
    }

    let [user, domain] = cleaned.split("@");

    return [user, domain]
}

export async function actorFromRemote(domain?: string, username?: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const { actor, followers, following } = await fetchRemoteActor(domain, username, f);

    const newActor: Actor = {
        domain: domain,
        follower: actor.followers?.toString() ?? "",
        inbox: actor.inbox?.toString(),
        iri: actor.id!.toString(),
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

    return { actor: newActor, remote: actor }
}

async function fetchRemoteActor(domain?: string, username?: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {

    let followers: APOrderedCollection | undefined;
    let following: APOrderedCollection | undefined;

    const webfingerURI = `https://${domain}/.well-known/webfinger?resource=acct:${username}@${domain}`;
    const webfingerRequest = await f(webfingerURI, { method: "GET" });

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
    const actorRequest = await f(actorURI.href, { method: "GET", headers: headers });

    if (!actorRequest.ok) {
        const errorResponse = await actorRequest.json();
        throw new ClientResponseError({ status: 500, response: errorResponse });
    }
    const apActor: APActor & { publicKey: { id: string, owner: string, publicKeyPem: string } } = await actorRequest.json();

    if (apActor.followers) {
        const followersRequest = await f(apActor.followers.toString(), { method: "GET", headers: headers });
        if (!followersRequest.ok) {
            const errorResponse = await webfingerRequest.json();
            throw new ClientResponseError({ status: 500, response: errorResponse });
        }
        followers = await followersRequest.json();
    }

    if (apActor.following) {
        const followingRequest = await f(apActor.following.toString(), { method: "GET", headers: headers });
        if (!followingRequest.ok) {
            const errorResponse = await webfingerRequest.json();
            throw new ClientResponseError({ status: 500, response: errorResponse });
        }
        following = await followingRequest.json();
    }



    return { actor: apActor, followers, following };
}
