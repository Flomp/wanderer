import { browser } from "$app/environment";
import { get, writable, type Writable } from "svelte/store";

type Theme = "dark" | "light"

export const theme: Writable<Theme> = writable(browser && localStorage.getItem("theme") as Theme || "light" );

export function toggleTheme() {
    const newTheme = get(theme) === "light" ? "dark" : "light";
    theme.set(newTheme)
    localStorage.setItem("theme", newTheme);
}