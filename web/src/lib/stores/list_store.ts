import { ExpandType, ExpandTypeToString, List, type ListFilter } from "$lib/models/list";
import type { Trail } from "$lib/models/trail";
import { APIError } from "$lib/util/api_util";
import type { Hits } from "meilisearch";
import { type AuthRecord, type ListResult } from "pocketbase";
import { get, writable, type Writable } from "svelte/store";
import { fetchGPX } from "./trail_store";
import { currentUser } from "./user_store";

let lists: List[] = []
export const list: Writable<List | null> = writable(null)
export const listTrail: Writable<Trail | null> = writable(null);

export async function lists_index(filter?: ListFilter, page: number = 1, perPage: number = 5,
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
    e: ExpandType = ExpandType.All) {
    const user = get(currentUser)
    
    const filterText = filter ? buildFilterText(user, filter) : ""

    const r = await f('/api/v1/list?' + new URLSearchParams({
        sort: `${filter?.sortOrder ?? "-"}${filter?.sort ?? "name"}`,
        perPage: perPage.toString(),
        page: page.toString(),
        filter: filterText,
        expand: ExpandTypeToString(e),
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const fetchedLists: ListResult<List> = await r.json();

    const result = page > 1 ? [...lists, ...fetchedLists.items] : fetchedLists.items

    lists = result;

    return { ...fetchedLists, items: result };
}


export async function lists_search_filter(filter: ListFilter, page: number = 1, perPage: number = 5, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch): Promise<ListResult<List>> {
    const user = get(currentUser)

    const filterText = buildSearchFilterText(user, filter)

    let r = await f("/api/v1/search/lists", {
        method: "POST",
        body: JSON.stringify({
            q: filter.q,
            options: {
                filter: filterText, sort: filter.sort && filter.sortOrder ? [`${filter.sort}:${filter.sortOrder == "+" ? "asc" : "desc"}`] : [],
                hitsPerPage: perPage,
                page: page
            }
        }),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const searchResult: { page: number, totalPages: number, hits: Hits<Record<string, any>> } = await r.json();


    const listIds = searchResult.hits.map((h: Record<string, any>) => h.id);

    if (listIds.length == 0) {
        return { items: [], page: searchResult.page, perPage, totalItems: 0, totalPages: searchResult.totalPages };
    }

    r = await f('/api/v1/list?' + new URLSearchParams({
        expand: "trails,trails.waypoints,trails.category,list_share_via_list",
        filter: `'${listIds.join(',')}'~id`,
        sort: filter.sort && filter.sortOrder ? `${filter.sortOrder}${filter.sort}` : ''
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const fetchedLists: ListResult<List> = await r.json();

    const result = page > 1 ? [...lists, ...fetchedLists.items] : fetchedLists.items

    lists = result;

    return { ...fetchedLists, items: result };

}

export async function lists_show(id: string, f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch
    , e: ExpandType = ExpandType.All) {

    const r = await f(`/api/v1/list/${id}?` + new URLSearchParams({
        expand: ExpandTypeToString(e)
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const response = await r.json()

    for (const trail of response.expand?.trails ?? []) {
        const gpxData: string = await fetchGPX(trail, f);
        trail.expand.gpx_data = gpxData;
    }




    list.set(response);

    return response;

}

export async function lists_create(list: List, avatar?: File) {
    const user = get(currentUser)
    if (!user) {
        throw Error("Unauthenticated")
    }
    list.author = user.id;

    let r = await fetch('/api/v1/list', {
        method: 'PUT',
        body: JSON.stringify(list),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
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
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    return await r.json();
}

export async function lists_update(list: List, avatar?: File) {
    let r = await fetch('/api/v1/list/' + list.id, {
        method: 'POST',
        body: JSON.stringify(list),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
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
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
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
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const model: List = await r.json();

    return model;
}

export async function lists_remove_trail(list: List, trail: Trail) {
    const r = await fetch('/api/v1/list/' + list.id, {
        method: 'POST',
        body: JSON.stringify({
            "trails-": trail.id
        }),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const model: List = await r.json();

    return model;
}

export async function lists_delete(list: List) {
    const r = await fetch('/api/v1/list/' + list.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }
}


function buildFilterText(user: AuthRecord | undefined, filter: ListFilter): string {
    let filterText = `(name~"${filter.q}"||description~"${filter.q}")`
    if (filter.author?.length) {
        filterText += `&&author="${filter.author}"`
    }
    if (user) {
        if (filter.public === false && filter.shared === false) {
            filterText += `&&author="${user.id}"`
        } else if (filter.public === true && filter.shared === false) {
            filterText += `&&(public=true||list_share_via_list.user!="${user.id}"||author="${user.id}")`
        } else if (filter.public === false && filter.shared === true) {
            filterText += `&&(public=false||list_share_via_list.user="${user.id}"||author="${user.id}")`
        }
    }
    return filterText
}

function buildSearchFilterText(user: AuthRecord, filter: ListFilter): string {
    let filterText: string = "";

    if (filter.author?.length) {
        filterText += `author = ${filter.author}`
    }

    if (filter.public !== undefined || filter.shared !== undefined) {
        if (filterText.length) {
            filterText += " AND "
        }
        filterText += "("
        if (filter.public !== undefined) {
            filterText += `(public = ${filter.public}`

            if (!filter.author?.length || filter.author == user?.id) {
                filterText += ` OR author = ${user?.id}`
            }
            filterText += ")"
        }

        if (filter.shared !== undefined) {
            if (filter.shared === true) {
                filterText += ` OR shares = ${user?.id}`
            } else {
                filterText += ` AND NOT shares = ${user?.id}`

            }
        }
        filterText += ")"
    }

    return filterText
}