import type { User } from "./user";

export class ListShare {
    id?: string;
    user: string;
    list: string;
    permission: "view" | "edit"
    expand?: {
        user: User
    }

    constructor(user: string, list: string, permission: "view" | "edit") {
        this.user = user;
        this.list = list;
        this.permission = permission
    }
}