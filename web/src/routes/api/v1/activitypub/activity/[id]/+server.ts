import { Collection, handleError, show } from "$lib/util/api_util";
import { json, type RequestEvent } from "@sveltejs/kit";
import type { APActivity, APOrderedCollectionPage, APRoot } from 'activitypub-types';
import type { Activity } from '$lib/models/activitypub/activity';

export async function GET(event: RequestEvent) {
    try {
        const a = await show<Activity>(event, Collection.activitypub_activities)

        const activity: APRoot<APActivity> = {
            id: a.iri,
            actor: a.actor,
            to: a.to,
            cc: a.cc,
            published: a.published,
            object: a.object

        }

        const headers = new Headers()
        headers.append("Content-Type", "application/activity+json")

        return json(activity, { headers })
    } catch (e: any) {
        return handleError(e)
    }
}