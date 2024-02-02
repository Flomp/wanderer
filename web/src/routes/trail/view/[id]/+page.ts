import { trails_show } from "$lib/stores/trail_store";
import type { Load } from "@sveltejs/kit";

export const load: Load = async ({ params }) => {
    try {
        await trails_show(params.id!, true)
    }catch(e) {
        
    }
};