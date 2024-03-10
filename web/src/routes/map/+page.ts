import type { TrailFilter } from "$lib/models/trail";
import { categories_index } from "$lib/stores/category_store";
import { trails } from "$lib/stores/trail_store";
import type { ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ params, locals, fetch }) => {
    const filter: TrailFilter = {
        q: "",
        category: [],
        difficulty: ["easy", "moderate", "difficult"],
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

    await categories_index(fetch)

    trails.set([])

    return { filter: filter }
};