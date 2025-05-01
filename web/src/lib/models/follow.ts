import type { Actor } from "./activitypub/actor";

export class Follow {
    id?: string;
    follower: string;
    followee: string;
    expand?: {
        follower: Actor,
        followee: Actor
    }

    constructor(follower: string, followee: string) {
        this.follower = follower;
        this.followee = followee;
    }
}