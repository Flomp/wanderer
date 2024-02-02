import { writable, type Writable } from "svelte/store";

export type Toast = {
    type: "info" | "success" | "warning" | "error"
    icon: string;
    text: string;
}


export const toast: Writable<Toast | null> = writable();


export function show_toast(newToast: Toast, duration: number = 3000) {
    toast.set(newToast);

    setTimeout(() => toast.set(null), duration);
}
