import type { User, UserAnonymous } from "./user";

export class Follow {
    id?: string;
    follower: string;
    followee: string;
    expand?: {
        follower: UserAnonymous,
        followee: UserAnonymous
    }

    constructor(follower: string, followee: string) {
        this.follower = follower;
        this.followee = followee;
    }
}