// Compare the rendered opening draft to the user's final draft. Defaults that
// merely appeared in the form are not persisted by a repair or a no-op save.
export function pluginSettingsConfigPatch(
    before: Record<string, unknown>, after: Record<string, unknown>,
): Record<string, unknown> {
    const patch: Record<string, unknown> = {};
    for (const key of new Set([...Object.keys(before), ...Object.keys(after)])) {
        if (same(before[key], after[key])) continue;
        if (key !== "categoryMapping" && object(before[key]) && object(after[key])) {
            patch[key] = pluginSettingsConfigPatch(before[key], after[key]);
        } else {
            patch[key] = after[key] === undefined ? null : after[key];
        }
    }
    return patch;
}

export function hasProviderConfigChanges(config?: Record<string, unknown>): boolean {
    if (!config) return false;
    const { host, ...other } = config;
    if (Object.keys(other).length) return true;
    if (!object(host)) return host !== undefined;
    return Object.keys(host).some(key => key !== "categoryRemapDismissedAt");
}

function object(value: unknown): value is Record<string, unknown> {
    return !!value && typeof value === "object" && !Array.isArray(value);
}

function same(a: unknown, b: unknown): boolean {
    if (object(a) && object(b)) {
        return Object.keys(a).length === Object.keys(b).length &&
            Object.keys(a).every(key => Object.hasOwn(b, key) && same(a[key], b[key]));
    }
    if (Array.isArray(a) && Array.isArray(b)) return a.length === b.length && a.every((item, i) => same(item, b[i]));
    return a === b;
}
