import { Trail } from "$lib/models/trail";
import { categories_index } from "$lib/stores/category_store";
import { trails_show } from "$lib/stores/trail_store";
import { error, type Load } from "@sveltejs/kit";

export const load: Load = async ({ params }) => {
    if (!params.id) {
        return error(400, "Bad Request")
    }
    const categories = await categories_index()

    let trail: Trail;
    if (params.id === "new") {
        trail = new Trail("", {category: categories[0]});
    } else {
        trail = await trails_show(params.id, true);
    }

    return { trail: trail }
};