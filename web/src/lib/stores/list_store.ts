import { List, type ListFilter } from "$lib/models/list";
import type { Trail } from "$lib/models/trail";
import { pb } from "$lib/pocketbase";
import { getFileURL } from "$lib/util/file_util";
import { ClientResponseError } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const lists: Writable<List[]> = writable([])
export const list: Writable<List> = writable(new List("", []))


export async function lists_index(filter?: ListFilter, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f('/api/v1/list?' + new URLSearchParams({
        sort: `${filter?.sortOrder ?? "-"}${filter?.sort ?? "name"}`,
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const fetchedLists: List[] = await r.json();

    lists.set(fetchedLists);

    if (fetchedLists.length > 0) {
        list.set(fetchedLists[0])
    }

    return fetchedLists;

}

export async function lists_create(list: List, avatar?: File) {
    if (!pb.authStore.model) {
        throw new Error("Unauthenticated");
    }

    list.author = pb.authStore.model!.id;

    let r = await fetch('/api/v1/list', {
        method: 'PUT',
        body: JSON.stringify(list),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const model: List = await r.json();

    const formData = new FormData();

    if (avatar) {
        formData.append("avatar", avatar);
    }

    r = await fetch(`/api/v1/list/${model.id!}/file`, {
        method: 'POST',
        body: formData,
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}

export async function lists_update(list: List, avatar?: File) {
    let r = await fetch('/api/v1/list/' + list.id, {
        method: 'POST',
        body: JSON.stringify(list),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const model: List = await r.json();

    const formData = new FormData();

    if (avatar) {
        formData.append("avatar", avatar);
    }

    r = await fetch(`/api/v1/list/${model.id!}/file`, {
        method: 'POST',
        body: formData,
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}

export async function lists_add_trail(list: List, trail: Trail) {
    const r = await fetch('/api/v1/list/' + list.id, {
        method: 'POST',
        body: JSON.stringify({
            "trails+": trail.id
        }),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}

export async function lists_remove_trail(list: List, trail: Trail) {
    const r = await fetch('/api/v1/list/' + list.id, {
        method: 'POST',
        body: JSON.stringify({
            "trails-": trail.id
        }),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}

export async function lists_delete(list: List) {
    const r = await fetch('/api/v1/list/' + list.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}