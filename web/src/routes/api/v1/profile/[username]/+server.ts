import { env } from "$env/dynamic/private";
import type { Actor } from "$lib/models/activitypub/actor";
import type { Profile } from '$lib/models/profile';
import { actorFromRemote, splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import { type APImage } from 'activitypub-types';

export async function GET(event: RequestEvent) {
    const fullUsername = event.params.username;
    if (!fullUsername) {
        return error(400, { message: "Bad request" })
    }
    const [username, domain] = splitUsername(fullUsername, new URL(env.ORIGIN).hostname)

    try {
        const actor: Actor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`domain='${domain}'&&username='${username}'`)

        const profile: Profile = {
            id: actor.id!,
            username: actor.username,
            acct: fullUsername,
            createdAt: actor.published ?? "",
            bio: actor.summary ?? "",
            uri: actor.iri,
            followers: actor.followerCount ?? 0,
            following: actor.followingCount ?? 0,
            icon: actor.icon ?? "",
        }

        return json(profile)
    } catch (e) {
        // actor does not exist yet
    }

    try {
        const { actor: newActor, remote: remoteActor } = await actorFromRemote(domain, username, event.fetch);

        const createdActor = await event.locals.pb.collection("activitypub_actors").create(newActor);

        const profile: Profile = {
            id: createdActor.id,
            username: remoteActor.name ?? remoteActor.preferredUsername ?? "",
            acct: `${username}@${domain}`,
            createdAt: remoteActor.published?.toString() ?? "",
            bio: remoteActor.summary ?? "",
            uri: remoteActor.id ?? "",
            followers: newActor.followerCount ?? 0,
            following: newActor.followingCount ?? 0,
            icon: (remoteActor.icon as APImage)?.url?.toString() ?? "",
        };

        return json(profile)
    } catch (e) {
        throw handleError(e)
    }
}
