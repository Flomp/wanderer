import { ExpandType } from "$lib/models/list";
import { Trail } from "$lib/models/trail";
import { pb } from "$lib/pocketbase";
import { categories_index } from "$lib/stores/category_store";
import { lists_index } from "$lib/stores/list_store";
import { trails_show } from "$lib/stores/trail_store";
import { error, type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {
    if (!params.id) {
        return error(400, "Bad Request")
    }
    const categories = await categories_index(fetch)
    const lists = await lists_index({ q: "", author: pb.authStore.record?.id ?? "" }, 1, -1, fetch, ExpandType.None)
    
    let trail: Trail;
    if (params.id === "new") {
        trail = new Trail("", { category: categories[0] });
    } else {
        trail = await trails_show(params.id, true, fetch);
    }

    return { trail: trail, lists: lists }
};