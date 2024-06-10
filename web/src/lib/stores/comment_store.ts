import { Comment } from "$lib/models/comment";
import type { Trail } from "$lib/models/trail";
import { pb } from "$lib/pocketbase";
import { ClientResponseError } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const comments: Writable<Comment[]> = writable([])

export async function comments_index(trail: Trail) {
    let r = await fetch('/api/v1/comment?' + new URLSearchParams({
        filter: `trail="${trail.id}"`,
    }), {
        method: 'GET',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }

    const fetchedComments: Comment[] = await r.json();

    comments.set(fetchedComments);

    return fetchedComments;
}

export async function comments_create(comment: Comment) {
    if (!pb.authStore.model) {
        throw new Error("Unauthenticated");
    }

    comment.author = pb.authStore.model!.id;

    let r = await fetch('/api/v1/comment', {
        method: 'PUT',
        body: JSON.stringify(comment),
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
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
        throw new ClientResponseError(await r.json())
    }

    const model: Comment = await r.json();

    return model;
}

export async function comments_delete(comment: Comment) {
    const r = await fetch('/api/v1/comment/' + comment.id, {
        method: 'DELETE',
    })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
}