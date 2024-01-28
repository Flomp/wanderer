import { categories_index } from "$lib/stores/category_store";
import type { Load } from "@sveltejs/kit";

export const load: Load = async ({ params }) => {
    await categories_index()
};