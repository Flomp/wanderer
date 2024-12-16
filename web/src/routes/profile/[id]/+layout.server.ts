import type { Follow } from "$lib/models/follow";
import type { Settings } from "$lib/models/settings";
import type { UserAnonymous } from "$lib/models/user";
import { follows_a_b, follows_counts, follows_index } from "$lib/stores/follow_store";
import { settings_show } from "$lib/stores/settings_store";
import { users_show } from "$lib/stores/user_store";
import { error, type ServerLoad } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

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
        if (e instanceof ClientResponseError && e.status == 404) {
            error(404, {
                message: 'Not found'
            });
        } else {
            throw e
        }
    }
};