import type { Trail } from "$lib/models/trail";
import { categories_index } from "$lib/stores/category_store";
import { feed_index } from "$lib/stores/feed_store";
import { trails_recommend } from "$lib/stores/trail_store";
import type { ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, fetch }) => {
    try {
        await categories_index(fetch)

        const feed = await feed_index(1, 10, fetch);

        const trails: Trail[] = await trails_recommend(4, fetch)
        return { trails: trails ?? [], feed }

    } catch (e) {
        console.error(e)
    }
    return { trails: [], feed: {items: [], page: 1, perPage: 1, totalItems: 0, totalPages: 0} }
};