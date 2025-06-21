import { follows_a_b } from "$lib/stores/follow_store";
import { profile_show } from "$lib/stores/profile_store";
import { APIError } from "$lib/util/api_util";
import { error, type NumericRange, type ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, fetch }) => {

    if (!params.handle) {
        error(404, "Not found")
    }

    try {
        const { actor, profile } = await profile_show(params.handle, fetch);

        const isOwnProfile = profile.id === locals.user.actor;

        const follow = await follows_a_b(locals.user.actor!, profile.id, fetch) ?? null
        return { profile, isOwnProfile, follow: follow, actor }

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