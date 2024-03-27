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
    language?: "en" | "de" | "nl" | "pl" | "pt";
    location?: { name: string, lat: number, lon: number }
}

export const currentUser: Writable<User | null> = writable<User | null>()

export async function users_create(user: User) {
    user.unit = "metric";
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
    } else {
        throw new ClientResponseError(await r.json())
    }

}

export async function logout() {
    pb.authStore.clear();
}

export async function users_update(user: User | { [K in keyof User]?: User[K] }, avatar?: File) {
    let r = await fetch('/api/v1/user/' + user.id, {
        method: 'POST',
        body: JSON.stringify(user)
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }


    const formData = new FormData();

    if (avatar) {
        formData.append("avatar", avatar);
    }

    r = await fetch(`/api/v1/user/${user.id!}/file`, {
        method: 'POST',
        body: formData,
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const model: User = await r.json();
    currentUser.set(model);
}

export async function users_delete(user: User) {
    const r = await fetch('/api/v1/user/' + user.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

}