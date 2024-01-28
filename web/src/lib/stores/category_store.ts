import { pb } from "$lib/constants";
import type { Category } from "$lib/models/category";
import { writable, type Writable } from "svelte/store";

export const categories: Writable<Category[]> = writable([])

export async function categories_index() {
    const response: Category[] = (await pb.collection('categories').getFullList<Category>())

    categories.set(response);
}