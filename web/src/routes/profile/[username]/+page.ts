import { activities_index } from "$lib/stores/activity_store";
import { lists_index } from "$lib/stores/list_store";
import { type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch, parent }) => {

    const { actor } = await parent()
    const lists = await lists_index({ q: "", author: params.id, sort: "created", sortOrder: "-" }, 1, -1, fetch)
    const activities = await activities_index(actor.iri + "/outbox", 1, 10, fetch);
    return { lists: lists.items, activities: activities }
};