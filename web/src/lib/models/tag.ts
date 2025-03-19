import type { Trail } from "./trail";

export class Tag {
    id?: string;
    name: string;

    constructor(name: string) {
        this.name = name;
    }
}

export class TrailTag {
    id?: string;
    tag: Tag;
    trail: Trail;

    constructor(tag: Tag, trail: Trail) {
        this.tag = tag;
        this.trail = trail;
    }
}