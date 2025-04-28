import type { UserAnonymous } from '$lib/models/user';
import { error, json, type RequestEvent } from '@sveltejs/kit';
import { ClientResponseError } from 'pocketbase';


export async function GET(event: RequestEvent) {

    let { username } = event.params;

    if (username?.startsWith('@')) {
        username = username.substring(1)
    }

    try {
        const user: UserAnonymous = await event.locals.pb.collection('users_anonymous').getFirstListItem(`username="${username}"`);
        const keys: { public_key: string } = await event.locals.pb.collection('activitypub').getFirstListItem(`user="${user.id}"`);

        return json({
            "@context": "https://www.w3.org/ns/activitystreams",
            "id": `${event.url.origin}/users/@${username}`,
            "type": "Person",
            "preferredUsername": username,
            "name": username,
            "summary": user.bio ?? '',
            "inbox": `${event.url.origin}/users/@${username}/inbox`,
            "outbox": `${event.url.origin}/users/@${username}/outbox`,
            "publicKey": {
                "id": `${event.url.origin}/users/@${username}#main-key`,
                "owner": `${event.url.origin}/users/@${username}`,
                "publicKeyPem": keys.public_key
            }
        });
    } catch (e) {
        if (e instanceof ClientResponseError) {
            return error(e.status, { message: e.message });
        }
        throw e;
    }


}

