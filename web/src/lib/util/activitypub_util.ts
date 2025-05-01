
export function splitUsername(username: string, localDomain: string) {
    const cleaned = username.replace(/^@/, "").trim();

    if (!cleaned.includes("@")) {
        return [username, undefined];
    }

    const [user, domain] = cleaned.split("@");

    return [user, domain]
}