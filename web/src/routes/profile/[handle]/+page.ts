import { lists_index } from "$lib/stores/list_store";
import { profile_timeline_index } from "$lib/stores/profile_store";
import { error, type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch, parent }) => {
    if (!params.handle) {
        error(404, "Not found")
    }

    const lists = await lists_index({ q: "", author: params.id, sort: "created", sortOrder: "-" }, 1, -1, fetch)
    const timeline = await profile_timeline_index(params.handle, 1, 10, fetch);
    return { lists: lists.items, timeline }
};