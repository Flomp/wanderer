import type { Trail } from "$lib/models/trail";
import { trails_show } from "$lib/stores/trail_store";

export const load = async ({ params, fetch }) => {
    const t: Trail = await trails_show("yesm2tqc6jok8jq", true, fetch)
    

    return { trail: t }
};