import { Comment } from "$lib/models/comment";
import type { Trail } from "$lib/models/trail";
import { pb } from "$lib/pocketbase";
import { APIError } from "$lib/util/api_util";
import { type ListResult } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const comments: Writable<Comment[]> = writable([])

export async function comments_index(trail: Trail) {
    let r = await fetch('/api/v1/comment?' + new URLSearchParams({
        filter: `trail="${trail.id}"`,
        sort: "-created"
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const fetchedComments: ListResult<Comment> = await r.json();

    comments.set(fetchedComments.items);

    return fetchedComments.items;
}

export async function comments_create(comment: Comment) {
    if (!pb.authStore.record) {
        throw new Error("Unauthenticated");
    }

    comment.author = pb.authStore.record!.id;

    let r = await fetch('/api/v1/comment', {
        method: 'PUT',
        body: JSON.stringify(comment),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const model: Comment = await r.json();

    return model;
}

export async function comments_update(comment: Comment) {
    let r = await fetch('/api/v1/comment/' + comment.id, {
        method: 'POST',
        body: JSON.stringify(comment),
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }

    const model: Comment = await r.json();

    return model;
}

export async function comments_delete(comment: Comment) {
    const r = await fetch('/api/v1/comment/' + comment.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail)
    }
}