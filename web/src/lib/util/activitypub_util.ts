

export function splitUsername(handle: string, localDomain?: string) {
    const cleaned = handle.replace(/^@/, "").trim();

    if (!cleaned.includes("@")) {
        return [cleaned, localDomain];
    }

    let [user, domain] = cleaned.split("@");

    return [user, domain]
}
