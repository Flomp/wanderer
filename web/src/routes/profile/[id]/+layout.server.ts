import { users_show } from "$lib/stores/user_store";
import { error, type ServerLoad } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export const load: ServerLoad = async ({ params, locals, fetch }) => {

    if (!params.id) {
        error(404, "Not found")
    }

    try {
        const user = await users_show(params.id, fetch);

        const isOwnProfile = params.id === locals.user?.id;

        return { user, isOwnProfile }
    } catch (e) {
        if (e instanceof ClientResponseError && e.status == 404) {
            error(404, {
                message: 'Not found'
            });
        }
        throw e
    }

};