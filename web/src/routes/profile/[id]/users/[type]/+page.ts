import { follows_index } from "$lib/stores/follow_store";
import { error, type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {
    if (!params.type || (params.type !== "followers" && params.type !== "following")) {
        throw error(400, "Missing or wrong type");
    }

    const follows = await follows_index({ [params.type == "followers" ? "followee" : "follower"]: params.id }, 1, 10, fetch)
    return { follows }
};