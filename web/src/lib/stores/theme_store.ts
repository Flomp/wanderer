import { browser } from "$app/environment";
import { get, writable, type Writable } from "svelte/store";

type Theme = "dark" | "light"

export const theme: Writable<Theme> = writable(browser && localStorage.getItem("theme") as Theme || "light" );

export function toggleTheme() {
    const currentTheme = get(theme);
    const newTheme = currentTheme === "light" ? "dark" : "light";
    document.documentElement.classList.remove(currentTheme)
    document.documentElement.classList.add(newTheme)    
    theme.set(newTheme)
    localStorage.setItem("theme", newTheme);
}