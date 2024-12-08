import { users_show } from "$lib/stores/user_store";
import { error, type ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, fetch }) => {

    if(!params.id) {
        error(404, "Not found")
    }

    const user = await users_show(params.id, fetch);

    const isOwnProfile = params.id === locals.user?.id;

    return { user, isOwnProfile }
};