import type { TrailFilter } from "$lib/models/trail";
import { trails_index, trails_search_filter } from "$lib/stores/trail_store";
import { type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {
    const filter: TrailFilter = {
        q: "",
        category: [],
        tags: [],
        difficulty: ["easy", "moderate", "difficult"],
        author: params.id,
        public: true,
        shared: true,
        near: {
            radius: 2000,
        },
        distanceMin: 0,
        distanceMax: 0,
        distanceLimit: 0,
        elevationGainMin: 0,
        elevationGainMax: 0,
        elevationGainLimit: 0,
        elevationLossMin: 0,
        elevationLossMax: 0,
        elevationLossLimit: 0,
        sort: "created",
        sortOrder: "+",
    };
    const trails = await trails_search_filter(filter, 1, fetch)
    return { trails, filter }
};