import type { DateTime } from "activitypub-types";

export interface Profile {
    id: string;
    username: string;
    preferredUsername: string;
    acct: string;
    bio: string;
    followers: number;
    following: number;
    uri: string;
    icon: string;
    createdAt: DateTime;
}