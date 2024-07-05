import { categories_index } from "$lib/stores/category_store";
import { type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {
    const categories = await categories_index(fetch)

    return { categories: categories }
};