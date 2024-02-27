import { object, string } from "yup";
import type { Trail } from "./trail";

export class List {
    id?: string;
    name: string;
    description?: string;
    avatar?: string;
    trails?: string[];
    expand?: {
        trails: Trail[]
    }
    author?: string;

    constructor(name: string, trails: Trail[], params?: { description?: string, avatar?: string, author?: string }) {
        this.name = name;
        this.expand = { trails: trails };
        this.description = params?.description;
        this.avatar = params?.description;
        this.author = params?.author;
    }
}

export interface ListFilter {
    q: string,
    sort: "name" | "size" | "created";
    sortOrder: "+" | "-"
}

export const listSchema = object<List>({
    id: string().optional(),
    name: string().required(),
    description: string().optional(),
    avatar: string().optional()
});