import type { Actor } from "./activitypub/actor";
import type { Trail } from "./trail";

export class TrailLike {
    id?: string;
    actor: string;
    trail: string;
    expand?: {
        actor: Actor,
        trail?: Trail
    }

    constructor(actor: string, trail: string) {
        this.actor = actor;
        this.trail = trail;
    }
}