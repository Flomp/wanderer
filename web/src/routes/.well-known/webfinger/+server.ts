import { error, json, RequestEvent } from '@sveltejs/kit';
import type PocketBase from "PocketBase";
// @ts-ignore
import { env } from "$env/dynamic/private"
// @ts-ignore
import { handleError } from "$lib/util/api_util";
// @ts-ignore
import { splitUsername } from "$lib/util/activitypub_util";
// @ts-ignore
import type { WebfingerResponse } from '$lib/models/activitypub/webfinger_response';

export async function GET(event: RequestEvent) {
    try {
        const resource = event.url.searchParams.get("resource")

        if (!resource || !resource.startsWith("acct:")) {
            return error(400, "Bad request");
        }

        const hostname = new URL(env.ORIGIN).hostname;
        const [username, domain] = splitUsername(resource.replace("acct:", ""))

        if (hostname !== domain) {
            return error(404, "Not found");
        }

        const id = `${env.ORIGIN}/api/v1/activitypub/user/${username}`

        const r: WebfingerResponse = {
            subject: resource,
            aliases: [id],
            links: [
                {
                    rel: 'self',
                    type: 'application/activity+json',
                    href: id,
                },
                {
                    rel: 'http://webfinger.net/rel/profile-page',
                    type: 'text/html',
                    href: `${env.ORIGIN}/profile/@${username}`,
                },
            ],
        }

        return json(r)
    } catch (err) {
        handleError(err)
    }
}