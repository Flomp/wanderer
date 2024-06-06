import type { User } from "./user";

export class TrailShare {
    id?: string;
    user: string;
    trail: string;
    permission: "view" | "edit"
    expand?: {
        user: User
    }

    constructor(user: string, trail: string, permission: "view" | "edit") {
        this.user = user;
        this.trail = trail;
        this.permission = permission
    }
}