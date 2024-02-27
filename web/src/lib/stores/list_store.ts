import { List, type ListFilter } from "$lib/models/list";
import type { Trail } from "$lib/models/trail";
import { pb } from "$lib/pocketbase";
import { getFileURL } from "$lib/util/file_util";
import { writable, type Writable } from "svelte/store";

export const lists: Writable<List[]> = writable([])
export const list: Writable<List> = writable(new List("", []))


export async function lists_index(filter?: ListFilter) {
    const dbResponse: List[] = (await pb.collection('lists').getFullList<List>({
        expand: "trails",
        sort: `${filter?.sortOrder ?? "+"}${filter?.sort ?? "name"}`
    }))

    for (const list of dbResponse) {
        list.avatar = getFileURL(list, list.avatar);
    }

    lists.set(dbResponse);

    if (dbResponse.length > 0) {
        list.set(dbResponse[0])
    }

    return dbResponse;
}

export async function lists_create(formData: { [key: string]: any; } | FormData) {
    if (!pb.authStore.model) {
        throw new Error("Unauthenticated");
    }

    formData.append("author", pb.authStore.model!.id);

    let model = await pb
        .collection("lists")
        .create<List>(formData);
}

export async function lists_update(list: List, formData: { [key: string]: any; } | FormData) {
    if ((formData.get("avatar") as File).size == 0) {
        formData.delete("avatar");
    }
    let model = await pb
        .collection("lists")
        .update<List>(list.id!, formData);
}

export async function lists_delete(list: List) {
    let success = await pb
        .collection("lists")
        .delete(list.id!);
}

export async function lists_add_trail(list: List, trail: Trail) {
    let model = await pb
        .collection("lists")
        .update(list.id!, { "trails+": trail.id });
}

export async function lists_remove_trail(list: List, trail: Trail) {
    let model = await pb
        .collection("lists")
        .update(list.id!, { "trails-": trail.id });
}