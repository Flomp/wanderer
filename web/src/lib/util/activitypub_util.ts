

export function splitUsername(handle: string, localDomain?: string) {
    const cleaned = handle.replace(/^@/, "").trim();

    if (!cleaned.includes("@")) {
        return [cleaned, localDomain];
    }

    let [user, domain] = cleaned.split("@");

    return [user, domain]
}

export function handleFromRecordWithIRI(record: any) {
    if (!record.expand?.author) {
        throw new Error("object has no author info")
    }
    
    if (!record.iri) {
        return `@${record.expand.author.username}`
    }
    const url = new URL(record.iri ?? "")

    return `@${record.expand.author.username}@${url.hostname}`
}