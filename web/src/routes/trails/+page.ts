import type { TrailFilter } from "$lib/models/trail";
import { categories_index } from "$lib/stores/category_store";
import { trails_search_filter } from "$lib/stores/trail_store";
import type { ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, url, fetch }) => {
    const filter: TrailFilter = {
        q: "",
        category: [],
        near: {
            radius: 2000,
        },
        distanceMin: 0,
        distanceMax: 20000,
        elevationGainMin: 0,
        elevationGainMax: 4000,
        sort: "created",
        sortOrder: "+",
    };
    const paramCategory = url.searchParams.get("category");
    if (paramCategory) {
        filter.category.push(paramCategory);
    }
    await trails_search_filter(filter, fetch);
    await categories_index()

    return {filter};
};