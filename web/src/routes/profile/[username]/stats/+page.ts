import type { SummitLogFilter } from "$lib/models/summit_log";
import { categories_index } from "$lib/stores/category_store";
import { summit_logs_index } from "$lib/stores/summit_log_store";
import { error, type Load } from "@sveltejs/kit";

export const load: Load = async ({ params, fetch }) => {

    if(!params.id) {
        error(404, "Not found")
    }

    const date = new Date()
    date.setUTCHours(6)
    const y = date.getFullYear()
    const m = date.getMonth();
    const firstDay = new Date(y, m, 2);
    const lastDay = new Date(y, m + 1, 1);

    const categories = await categories_index(fetch)

    const filter: SummitLogFilter = {
        startDate: firstDay.toISOString().slice(0, 10),
        endDate: lastDay.toISOString().slice(0, 10),
        category: []
    }
    const logs = await summit_logs_index(params.id, filter, fetch);

    return { filter }
};