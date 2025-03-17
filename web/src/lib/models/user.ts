import type { Settings } from "./settings";
import { type AuthRecord } from "pocketbase";

export type User = AuthRecord & {
    id: string,
    username?: string,
    email?: string,
    password: string,
    avatar?: string;
    bio?: string;
    created?: string;
}

export type UserAnonymous = {
    id: string,
    username?: string,
    avatar?: string;
    created?: string;
    private: boolean;
}