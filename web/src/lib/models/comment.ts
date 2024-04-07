import type { User } from "$lib/stores/user_store";

export class Comment {
    id?: string;
    text: string;
    rating: number;
    author: string;
    trail: string;
    created?: string;
    expand?: {
        author: User
    }

    constructor(text: string, rating: number, author: string, trail: string) {
        this.text = text;
        this.rating = rating;
        this.author = author;
        this.trail = trail;
    }
}