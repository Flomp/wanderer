import { List } from "$lib/models/list";
import { lists_show } from "$lib/stores/list_store";
import { error, type Load } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export const load: Load = async ({ params, fetch, data }) => {
    if (!params.id) {
        return error(400, "Bad Request")
    }

    let list: List;
    if (params.id === "new") {
        list = new List("", []);
        return { list: list }
    } else {
        try {
            list = await lists_show(params.id, fetch);
            return { list: list }

        } catch (e) {
            if (e instanceof ClientResponseError) {
                return error(e.status as any, e.message)
            }
        }
    }

};