import { List, type ListFilter } from "$lib/models/list";
import type { Trail } from "$lib/models/trail";
import { pb } from "$lib/pocketbase";
import { ClientResponseError, type ListResult } from "pocketbase";
import { get, writable, type Writable } from "svelte/store";
import { fetchGPX } from "./trail_store";

let lists: List[] = []
export const list: Writable<List | null> = writable(null)
export const listTrail: Writable<Trail | null> = writable(null);

export async function lists_index(filter?: ListFilter, page: number = 1, perPage: number = 5, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const filterText = filter ? buildFilterText(filter) : ""

    const r = await f('/api/v1/list?' + new URLSearchParams({
        sort: `${filter?.sortOrder ?? "-"}${filter?.sort ?? "name"}`,
        "per-page": perPage.toString(),
        page: page.toString(),
        filter: filterText
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const fetchedLists: ListResult<List> = await r.json();

    const result = page > 1 ? [...lists, ...fetchedLists.items] : fetchedLists.items

    lists = result;

    return { ...fetchedLists, items: result };

}

export async function lists_show(id: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch) {
    const r = await f(`/api/v1/list/${id}`, {
        method: 'GET',
    })
    const response = await r.json()

    for (const trail of response.expand?.trails ?? []) {
        const gpxData: string = await fetchGPX(trail, f);
        trail.expand.gpx_data = gpxData;
    }


    if (!r.ok) {
        throw new ClientResponseError(response)
    }

    list.set(response);

    return response;

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

    if (r.ok) {
        return await r.json()
    } else {
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

function buildFilterText(filter: ListFilter): string {


    let filterText = `(name~"${filter.q}"||description~"${filter.q}")`



    if (filter.author?.length) {
        filterText += `&&author="${filter.author}"`
    }

    if (pb.authStore.model) {
        if (filter.public === false && filter.shared === false) {
            filterText += `&&author="${pb.authStore.model.id}"`
        } else if (filter.public === true && filter.shared === false) {
            filterText += `&&(public=true||list_share_via_list.user!="${pb.authStore.model.id}"||author="${pb.authStore.model.id}")`
        } else if (filter.public === false && filter.shared === true) {
            filterText += `&&(public=false||list_share_via_list.user="${pb.authStore.model.id}"||author="${pb.authStore.model.id}")`
        }
    }

    return filterText
}