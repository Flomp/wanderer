import type { User } from "../user";

export interface Actor {
    username: string;
    domain?: string;
    summary?: string;
    published?: string;
    followerCount?: number,
    followingCount?: number,
    IRI: string;
    inbox: string;
    outbox?: string;
    icon?: string;
    follower?: string;
    following?: string;
    isLocal: boolean;
    public_key: string;
    last_fetched: string;
    user?: string
    expand?: {
        user?: User
    }
}