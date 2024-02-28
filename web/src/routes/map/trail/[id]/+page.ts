import { trails, trails_show } from "$lib/stores/trail_store";
import { error, type ServerLoad } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export const load: ServerLoad = async ({ params, locals }) => {
    try {
        await trails_show(params.id!, true)
    } catch (e) {
        if (e instanceof ClientResponseError && e.status == 404) {
            error(404, {
                message: 'Not found'
            });
        }

    }
};