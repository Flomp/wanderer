import type { Settings } from "./settings";

export type User = {
    id: string,
    username?: string,
    email?: string,
    password: string,
    avatar?: string;
    language?: string;
}