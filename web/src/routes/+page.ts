import { categories_index } from "$lib/stores/category_store";
import { trails_index } from "$lib/stores/trail_store";
import type { Load } from "@sveltejs/kit";

export const load: Load = async ({ params }) => {
    await trails_index()
    await categories_index()
};