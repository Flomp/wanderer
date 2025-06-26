import type { User } from "../user";

export interface Actor {
    id?: string;
    username: string;
    preferred_username: string;
    domain?: string;
    summary?: string;
    published?: string;
    followerCount?: number,
    followingCount?: number,
    iri: string;
    inbox: string;
    outbox?: string;
    icon?: string;
    followers?: string;
    following?: string;
    isLocal: boolean;
    public_key: string;
    last_fetched: string;
    user?: string
    expand?: {
        user?: User
    }
}