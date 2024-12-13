import type { List, ListFilter } from "$lib/models/list";
import { lists_index, lists_show } from "$lib/stores/list_store";
import { error, type Load } from "@sveltejs/kit";
import { ClientResponseError, type ListResult } from "pocketbase";

export const load: Load = async ({ params, fetch, url }) => {
    const filter: ListFilter = {
        q: "",
        author: "",
        shared: true,
        public: true,
        sort: "created",
        sortOrder: "+",
    };

    let lists: ListResult<List>;
    if (url.searchParams.get("list")) {
        try {
            const list = await lists_show(url.searchParams.get("list") ?? "", fetch)

            lists = { items: [list], page: 1, perPage: 1, totalItems: 1, totalPages: 1 }
        } catch (e) {
            if (e instanceof ClientResponseError && e.status == 404) {
                error(404, {
                    message: 'Not found'
                });
            }
            throw e
        }
    } else {
        lists = await lists_index(filter, 1, undefined, fetch);
    }

    return { lists, filter }
};