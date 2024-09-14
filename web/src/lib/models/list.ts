import { object, string } from "yup";
import type { Trail } from "./trail";
import type { ListShare } from "./list_share";

export class List {
    id?: string;
    name: string;
    description?: string;
    avatar?: string;
    trails?: string[];
    expand?: {
        trails?: Trail[]
        list_share_via_list?: ListShare[]

    }
    author?: string;

    constructor(name: string, trails: Trail[], params?: { description?: string, avatar?: string, author?: string }) {
        this.name = name;
        this.expand = { trails: trails };
        this.trails = trails.map(t => t.id!);
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