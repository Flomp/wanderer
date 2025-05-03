import type { Actor } from "$lib/models/activitypub/actor";
import { follows_a_b, follows_counts } from "$lib/stores/follow_store";
import { profile_show } from "$lib/stores/profile_store";
import { users_show } from "$lib/stores/user_store";
import { APIError } from "$lib/util/api_util";
import { error, type NumericRange, type ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, fetch }) => {

    if (!params.username) {
        error(404, "Not found")
    }

    try {
        const profile = await profile_show(params.username, fetch);
        const actor: Actor = await locals.pb.collection("activitypub_actors").getFirstListItem(`user = '${locals.user.id}'`)

        const isOwnProfile = profile.id === actor.id;

        const follow = await follows_a_b(actor.id!, profile.id, fetch) ?? null
        return { profile, isOwnProfile, follow: follow }

    } catch (e) {
        if (e instanceof APIError) {
            error(e.status as NumericRange<400, 599>, {
                message: e.status == 404 ? 'Not found' : e.message
            });
        } else {
            throw e
        }
    }
};