import { error, type ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, url, fetch }) => {
    if(!params.token) {
        error(400, "Invalid token")
    }
};