import type { Category } from "$lib/models/category";
import { pb } from "$lib/pocketbase";
import { error, json, type RequestEvent } from "@sveltejs/kit";

export async function GET(event: RequestEvent) {
    try {
        const r: Category[] = await pb.collection('categories').getFullList<Category>()
        return json(r)
    } catch (e: any) {
        throw error(e.status, e);
    }

}