import type { Actor } from "./activitypub/actor";
import type { User } from "./user";

export class ListShare {
    id?: string;
    actor: string;
    list: string;
    permission: "view" | "edit"
    expand?: {
        actor: Actor
    }

    constructor(user: string, list: string, permission: "view" | "edit") {
        this.actor = user;
        this.list = list;
        this.permission = permission
    }
}