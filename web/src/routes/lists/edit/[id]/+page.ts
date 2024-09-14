import { List } from "$lib/models/list";
import { lists_show } from "$lib/stores/list_store";
import { getFileURL } from "$lib/util/file_util";
import { error, type Load } from "@sveltejs/kit";
import { ClientResponseError } from "pocketbase";

export const load: Load = async ({ params, fetch, data }) => {
    if (!params.id) {
        return error(400, "Bad Request")
    }

    let list: List;
    if (params.id === "new") {
        list = new List("", []);
        return { list: list, previewUrl: "" }
    } else {
        try {
            list = await lists_show(params.id, fetch);
            const previewURL = getFileURL(list, list.avatar);

            return { list: list, previewUrl: previewURL }

        } catch (e) {           
            if (e instanceof ClientResponseError) {
                return error(e.status as any, e.message)
            }
        }
    }

};