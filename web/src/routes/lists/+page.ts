import { ExpandType, type List, type ListFilter } from "$lib/models/list";
import { lists_index, lists_show } from "$lib/stores/list_store";
import { APIError } from "$lib/util/api_util";
import { error, type Load, type NumericRange } from "@sveltejs/kit";
import { type ListResult } from "pocketbase";

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
            const list = await lists_show(url.searchParams.get("list") ?? "", fetch, ExpandType.None)

            lists = { items: [list], page: 1, perPage: 1, totalItems: 1, totalPages: 1 }
        } catch (e) {
            if (e instanceof APIError) {
                error(e.status as NumericRange<400, 599>, {
                    message: e.status == 404 ? 'Not found' : e.message
                });
            }
            throw e
        }
    } else {
        lists = await lists_index(filter, 1, undefined, fetch);
    }

    return { lists, filter }
};