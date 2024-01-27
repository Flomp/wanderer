import type { Load } from "@sveltejs/kit";
import type {Trail} from "$lib/models/trail"
import { trails_show } from "$lib/stores/trail_store";

export const load: Load = async ({ params }) => {
    await trails_show(params.id!, true)
};