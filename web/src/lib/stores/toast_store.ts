import { writable, type Writable } from "svelte/store";

export type Toast = {
    type: "info" | "success" | "warning" | "error"
    icon: string;
    text: string;
}


export const toast: Writable<Toast> = writable();
