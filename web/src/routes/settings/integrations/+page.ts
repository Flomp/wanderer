import { integrations_index } from "$lib/stores/integration_store";
import { type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {
    const integrations = await integrations_index(fetch)
    return { integration: integrations.at(0) }
};