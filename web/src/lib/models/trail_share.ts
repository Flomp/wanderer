import type { Actor } from "./activitypub/actor";

export class TrailShare {
    id?: string;
    actor: string;
    trail: string;
    permission: "view" | "edit"
    expand?: {
        actor: Actor
    }

    constructor(actor: string, trail: string, permission: "view" | "edit") {
        this.actor = actor;
        this.trail = trail;
        this.permission = permission
    }
}