import type { Actor } from "./activitypub/actor";

export class Follow {
    id?: string;
    follower: string;
    followee: string;
    status: "pending" | "accepted";
    expand?: {
        follower: Actor,
        followee: Actor
    }

    constructor(follower: string, followee: string) {
        this.follower = follower;
        this.followee = followee;
        this.status = "pending"
    }
}