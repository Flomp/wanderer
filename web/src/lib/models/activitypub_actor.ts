import type { User } from "./user";

export interface ActivityPubActor {
    iri: string;
    username: string;
    domain: string;
    icon: string;
    inbox: string;
    outbox: string;
    isLocal: boolean;
    publicKey: string;
    user?: string
    expand?: {
        user?: User
    }
}