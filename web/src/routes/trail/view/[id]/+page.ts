import { lists_index } from "$lib/stores/list_store";
import { trails_show } from "$lib/stores/trail_store";
import type { Load } from "@sveltejs/kit";

export const load: Load = async ({ params }) => {
    await trails_show(params.id!, true)
    await lists_index();
};