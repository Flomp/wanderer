import type { Settings } from "./settings";

export type User = {
    id: string,
    username?: string,
    email?: string,
    password: string,
    avatar?: string;
    bio?: string;
    language?: string;
    created?: string;
}

export type UserAnonymous = {
    id: string,
    username?: string,
    avatar?: string;
    bio?: string;
    created?: string;
}