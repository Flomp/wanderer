import { categories_index } from "$lib/stores/category_store";
import { trails } from "$lib/stores/trail_store";
import type { ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals }) => {
    await categories_index()

    trails.set([])
};