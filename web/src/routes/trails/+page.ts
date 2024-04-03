import type { TrailFilter } from "$lib/models/trail";
import { categories_index } from "$lib/stores/category_store";
import { trails_get_filter_values, trails_search_filter } from "$lib/stores/trail_store";
import type { ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, url, fetch }) => {
    const filterValues = await trails_get_filter_values(fetch);

    const filter: TrailFilter = {
        q: "",
        category: [],
        difficulty: ["easy", "moderate", "difficult"],
        near: {
            radius: 2000,
        },
        distanceMin: 0,
        distanceMax: filterValues.max_distance,
        distanceLimit: filterValues.max_distance,
        elevationGainMin: 0,
        elevationGainMax: filterValues.max_elevation_gain,
        elevationGainLimit: filterValues.max_elevation_gain,
        sort: "created",
        sortOrder: "+",
    };
    const paramCategory = url.searchParams.get("category");
    if (paramCategory) {
        filter.category.push(paramCategory);
    }
    const response = await trails_search_filter(filter, 1, fetch);
    await categories_index(fetch)

    return {
        filter: filter, pagination: { page: response.page, totalPages: response.totalPages }
    };
};