import type { Trail } from "$lib/models/trail";


export function splitUsername(handle: string, localDomain?: string) {
    const cleaned = handle.replace(/^@/, "").trim();

    if (!cleaned.includes("@")) {
        return [cleaned, localDomain];
    }

    let [user, domain] = cleaned.split("@");

    return [user, domain]
}

export function handleFromTrail(trail: Trail) {
    if (!trail.expand?.author) {
        throw new Error("trail has no author info")
    }

    if (!trail.remote_url) {
        return `@${trail.expand.author.username}`
    }
    const url = new URL(trail.remote_url)

    return `@${trail.expand.author.username}@${url.hostname}`
}