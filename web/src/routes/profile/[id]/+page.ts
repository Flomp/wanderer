import { lists_index } from "$lib/stores/list_store";
import { type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {

    const lists = await lists_index({ q: "", author: params.id, sort: "created", sortOrder: "-" }, 1, -1, fetch)
    return { lists: lists.items }
};