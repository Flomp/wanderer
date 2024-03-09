import { regenerateInstance } from "$lib/meilisearch";
import { pb } from "$lib/pocketbase";
import { ClientResponseError } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export type User = {
    id: string,
    username?: string,
    email?: string,
    password: string,
    avatar?: string;
    unit?: "metric" | "imperial";
    language?: "en" | "de";
    location?: { name: string, lat: number, lon: number }
}

export const currentUser: Writable<User | null> = writable<User | null>()

export async function users_create(user: User) {
    const r = await fetch('/api/v1/user', {
        method: 'PUT',
        body: JSON.stringify({ ...user, passwordConfirm: user.password })
    })

    if (r.ok) {
        return await r.json();
    } else {
        throw new ClientResponseError(await r.json())
    }

}

export async function login(user: User) {
    const r = await fetch('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify(user),
    })

    if (r.ok) {
        pb.authStore.loadFromCookie(document.cookie)
        regenerateInstance()
    } else {
        throw new ClientResponseError(await r.json())
    }

}

export async function logout() {
    pb.authStore.clear();
    regenerateInstance()
}

export async function users_update(id: string, user: User | { [K in keyof User]?: User[K] } | FormData) {
    const r = await fetch('/api/v1/user', {
        method: 'POST',
        body: JSON.stringify({ id: id, user: user })
    })

    if (r.ok) {
        const model = await r.json();
        currentUser.set(model);
    } else {
        throw new ClientResponseError(await r.json())
    }
}

export async function users_delete(user: User) {
    const r = await fetch('/api/v1/user', {
        method: 'DELETE',
        body: JSON.stringify({ id: user.id })
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

}