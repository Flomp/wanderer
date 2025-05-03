import { env } from '$env/dynamic/private';
import type { Actor } from '$lib/models/activitypub/actor';
import { splitUsername } from '$lib/util/activitypub_util';
import { handleError } from '$lib/util/api_util';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import { type APActor } from 'activitypub-types';
import type { UserAnonymous } from "$lib/models/user";


export async function GET(event: RequestEvent) {

    try {
        let fullUsername = event.params.username;
        if (!fullUsername) {
            return error(400, "Bad request");
        }

        fullUsername = fullUsername.replace(/^@/, "");

        const [username, domain] = splitUsername(fullUsername, env.ORIGIN)

        const user: UserAnonymous = await event.locals.pb.collection("users_anonymous").getFirstListItem(`username='${username}'`)

        const actor: Actor = await event.locals.pb.collection("activitypub_actors").getFirstListItem(`user='${user.id}'`)

        const id = `${env.ORIGIN}/api/v1/activitypub/user/${username}`

        const r: APActor & { publicKey: { id: string, owner: string, publicKeyPem: string } } = {
            id: id,
            type: "Person",
            inbox: id + '/inbox',
            outbox: id + '/outbox',
            summary: user.bio,
            name: user.username,
            preferredUsername: user.username,
            followers: id + '/followers',
            following: id + '/following',
            url: id,
            published: user.created,
            icon: {
                type: "Image",
                url: `${env.ORIGIN}/api/v1/files/users/${user.id}/${user.avatar}`
            },
            publicKey: {
                id: id + '#main-key',
                owner: id,
                publicKeyPem: actor.public_key
            }

        }

        return json(r);
    } catch (e) {
        throw handleError(e)
    }


}

