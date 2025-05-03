import type { User } from "../user";

export interface Activity {
    id: string;
    iri: string;
    type: string;
    to: string[];
    cc: string[];
    bto: string[];
    bcc: string[];
    object: string;
    actor: string;
    published: string;
}