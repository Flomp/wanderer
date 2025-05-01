import type { DateTime } from "activitypub-types";

export interface Profile {
    username: string;
    acct: string;
    bio: string;
    followers: number;
    following: number;
    uri: string;
    icon: string;
    createdAt: DateTime;
}