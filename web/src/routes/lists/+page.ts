import type { TrailFilter } from "$lib/models/trail";
import { lists_index } from "$lib/stores/list_store";
import type { Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {
    const filter: TrailFilter = {
        q: "",
        category: [],
        difficulty: ["easy", "moderate", "difficult"],
        near: {
            radius: 2000,
        },
        distanceMin: 0,
        distanceMax: 20000,
        distanceLimit: 20000,
        elevationGainMin: 0,
        elevationGainMax: 4000,
        elevationGainLimit: 4000,
        sort: "created",
        sortOrder: "+",
    };
    await lists_index(undefined, fetch);

    return {filter};
};