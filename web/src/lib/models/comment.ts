import type { Actor } from "./activitypub/actor";

export class Comment {
    id?: string;
    text: string;
    rating?: number;
    author: string;
    trail: string;
    created?: string;
    updated?: string;
    expand?: {
        author: Actor
    }

    constructor(text: string, rating: number, author: string, trail: string) {
        this.text = text;
        this.rating = rating;
        this.author = author;
        this.trail = trail;
    }
}