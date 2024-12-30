import { object, string } from "yup";
import type { Trail } from "./trail";
import type { ListShare } from "./list_share";
import type { UserAnonymous } from "./user";

export class List {
    id?: string;
    name: string;
    public: boolean;
    description?: string;
    avatar?: string;
    trails?: string[];
    expand?: {
        trails?: Trail[]
        list_share_via_list?: ListShare[]
        author?: UserAnonymous;

    }
    author: string;

    constructor(name: string, trails: Trail[], params?: { description?: string, public?: boolean, avatar?: string, author?: string }) {
        this.name = name;
        this.public = params?.public ?? false
        this.expand = { trails: trails };
        this.trails = trails.map(t => t.id!);
        this.description = params?.description;
        this.avatar = params?.avatar;
        this.author = params?.author ?? "000000000000000";
    }
}

export interface ListFilter {
    q: string,
    sort?: "name" | "size" | "created";
    author?: string;
    public?: boolean;
    shared?: boolean;
    sortOrder?: "+" | "-"
}

export const listSchema = object<List>({
    id: string().optional(),
    name: string().required(),
    description: string().optional(),
    avatar: string().optional()
});