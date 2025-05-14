import type { TrailFilter } from "$lib/models/trail";
import { profile_trails_index } from "$lib/stores/profile_store";
import { error, type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch, parent }) => {
    if (!params.username) {
        error(404, "Not found")
    }
    const { actor } = await parent()

    const filter: TrailFilter = {
        q: "",
        category: [],
        tags: [],
        difficulty: ["easy", "moderate", "difficult"],
        author: actor.id,
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
    const trails = await profile_trails_index(params.username, filter, 1, 12, fetch)
    return { trails, filter }
};