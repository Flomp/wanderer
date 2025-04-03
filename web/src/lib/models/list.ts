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

export const enum ExpandType {
    None = 0,
    Trails = 1 << 0,
    Waypoints = 1 << 1,
    TrailCategories = 1 << 2,
    ListShares = 1 << 3,
    All = ~(~0 << 4),
}

export function ExpandTypeToString(e: ExpandType) : string {

    if (e == ExpandType.None)
        return "";

    var ret = "";
    if ((e & ExpandType.Trails) === ExpandType.Trails) {
        ret += "trails,";
    }
    if ((e & ExpandType.Waypoints) === ExpandType.Waypoints) {
        ret += "trails.waypoints,";
    }
    if ((e & ExpandType.TrailCategories) === ExpandType.TrailCategories) {
        ret += "trails.category,";
    }
    if ((e & ExpandType.ListShares) === ExpandType.ListShares) {
        ret += "list_share_via_list,";
    }

    return ret.slice(0, -1);
} 