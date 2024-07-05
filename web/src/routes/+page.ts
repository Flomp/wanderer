import { categories_index } from "$lib/stores/category_store";
import { trails_index } from "$lib/stores/trail_store";
import type { ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, fetch }) => {
    await trails_index(20, true, fetch)
    await categories_index(fetch)
};