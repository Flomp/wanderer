import type { Actor } from "./activitypub/actor";

export class Comment {
    id?: string;
    text: string;
    author: string;
    trail: string;
    created?: string;
    updated?: string;
    expand?: {
        author: Actor
    }

    constructor(text: string, author: string, trail: string) {
        this.text = text;
        this.author = author;
        this.trail = trail;
    }
}