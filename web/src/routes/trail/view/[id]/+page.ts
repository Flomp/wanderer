import { trails_show } from "$lib/stores/trail_store";
import { APIError } from "$lib/util/api_util";
import { error, type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {
    try {
        const trail = await trails_show(params.id!, true, fetch)

        return { trail }
    } catch (e) {
        if (e instanceof APIError && e.status == 404) {
            error(404, {
                message: 'Not found'
            });
        }
        console.log(e);
    }

};