import type { Profile } from '$lib/models/profile';
import { actorFromDb, splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';

export async function GET(event: RequestEvent) {
    const fullUsername = event.params.handle;
    if (!fullUsername) {
        return error(400, { message: "Bad request" })
    }

    try {
        const actor = await actorFromDb(event.locals.pb, fullUsername, event.fetch);


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

        return json({ profile, actor: actor })
    } catch (e) {
        if (e instanceof Error && e.message == "fetch failed") {
            return error(404, "Not found")
        }
        return handleError(e)
    }
}
