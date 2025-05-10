import { env } from "$env/dynamic/private";
import type { Profile } from '$lib/models/profile';
import { actorFromRemote, splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    const fullUsername = event.params.username;
    if (!fullUsername) {
        return error(400, { message: "Bad request" })
    }
    const [username, domain] = splitUsername(fullUsername)


    try {
        const { actor: fetchedActor, remote: remoteActor } = await actorFromRemote(domain, username, event.fetch);

        try {
            const dbActor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`iri='${fetchedActor.iri}'`)
            fetchedActor.id = dbActor.id
        } catch (e) {
            const dbActor = await event.locals.pb.collection("activitypub_actors").create(fetchedActor);
            fetchedActor.id = dbActor.id
        }

        const profile: Profile = {
            id: fetchedActor.id!,
            username: fetchedActor.username,
            acct: fullUsername,
            createdAt: fetchedActor.published ?? "",
            bio: fetchedActor.summary ?? "",
            uri: fetchedActor.iri,
            followers: fetchedActor.followerCount ?? 0,
            following: fetchedActor.followingCount ?? 0,
            icon: fetchedActor.icon ?? "",
        }

        return json({profile, actor: fetchedActor})
    } catch (e) {
        throw handleError(e)
    }
}
