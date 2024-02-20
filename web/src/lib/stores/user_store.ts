import { regenerateInstance } from "$lib/meilisearch";
import { pb } from "$lib/pocketbase";
import { type AuthModel } from 'pocketbase';
import { writable, type Writable } from "svelte/store";

export type User = {
    username?: string,
    email?: string,
    password: string,
}

export const currentUser: Writable<AuthModel | null> = writable<AuthModel | null>()

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