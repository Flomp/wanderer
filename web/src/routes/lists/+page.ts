import { lists_index } from "$lib/stores/list_store";
import type { Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch, url }) => {
    await lists_index(undefined, fetch);
};