import type { User, UserAnonymous } from "./user";

export class TrailShare {
    id?: string;
    user: string;
    trail: string;
    permission: "view" | "edit"
    expand?: {
        user: UserAnonymous
    }

    constructor(user: string, trail: string, permission: "view" | "edit") {
        this.user = user;
        this.trail = trail;
        this.permission = permission
    }
}