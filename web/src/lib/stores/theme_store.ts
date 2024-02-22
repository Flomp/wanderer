import { get, writable, type Writable } from "svelte/store";

type Theme = "dark" | "light"

export const theme: Writable<Theme> = writable("light");


export function toggleTheme() {
    if (get(theme) === "light") {
        theme.set("dark")
    } else {
        theme.set("light");
    }
}