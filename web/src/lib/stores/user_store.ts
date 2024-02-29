import { regenerateInstance } from "$lib/meilisearch";
import { pb } from "$lib/pocketbase";
import { writable, type Writable } from "svelte/store";

export type User = {
    id: string,
    username?: string,
    email?: string,
    password: string,
    avatar?: string;
    unit?: "metric" | "imperial";
    language?: "en" | "de";
    location?: {name: string, lat: number, lon: number}
}

export const currentUser: Writable<User | null> = writable<User | null>()

export async function users_create(user: User) {
    const model = await pb.collection('users').create({ ...user, passwordConfirm: user.password });

    return model;
}

export async function login(user: User) {
    const authData = await pb.collection('users').authWithPassword(user.email ?? user.username!, user.password);
    regenerateInstance()
}

export async function logout() {
    pb.authStore.clear();
    regenerateInstance()
}

export async function users_update(id: string, user: User | { [K in keyof User]?: User[K] } | FormData) {
    let model = await pb
        .collection("users")
        .update<User>(id, user);

    currentUser.set(model);
}

export async function users_delete(user: User) {
    let success = await pb
        .collection("users")
        .delete(user.id);
}