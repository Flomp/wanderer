import { follows_a_b, follows_counts } from "$lib/stores/follow_store";
import { users_show } from "$lib/stores/user_store";
import { APIError } from "$lib/util/api_util";
import { error, type NumericRange, type ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, fetch }) => {

    if (!params.id) {
        error(404, "Not found")
    }

    try {
        const user = await users_show(params.id, fetch);
        if(user.private && locals.user.id != params.id) {
            error(401, {
                message: "Profile is private"
            })
        }
        const isOwnProfile = params.id === locals.user?.id;

        const follow = await follows_a_b(locals.user?.id, params.id, fetch) ?? null
        const followCounts = await follows_counts(params.id, fetch)
        return { user, isOwnProfile, ...followCounts, follow: follow }

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