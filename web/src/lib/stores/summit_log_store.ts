import { pb } from "$lib/pocketbase";
import { SummitLog } from "$lib/models/summit_log";
import { writable, type Writable } from "svelte/store";

export const summitLog: Writable<SummitLog> = writable(new SummitLog(new Date().toISOString()));

export async function summit_logs_create(bodyParams?: { [key: string]: any; } | FormData) {
    const model = await pb
        .collection("summit_logs")
        .create<SummitLog>(bodyParams);

    return model;
}

export async function summit_logs_update(updatedSummitLog: SummitLog) {
    const model = await pb
        .collection("summit_logs")
        .update(updatedSummitLog.id!, updatedSummitLog);

    return model;
}

export async function summit_logs_delete(id: string) {
    const success = await pb
        .collection("summit_logs")
        .delete(id);

    return success;
}