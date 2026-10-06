// Mirror the host's MergePluginConfig: recursively merge JSON objects, except
// categoryMapping, where an empty or partial override deliberately replaces defaults.
export function mergePluginConfig(...layers: Record<string, unknown>[]): Record<string, unknown> {
    const result: Record<string, unknown> = {};
    for (const layer of layers) mergeInto(result, layer);
    return result;
}

function isObject(value: unknown): value is Record<string, unknown> {
    return value !== null && typeof value === "object" && !Array.isArray(value);
}

function clone(value: unknown): unknown {
    if (Array.isArray(value)) return value.map(clone);
    if (isObject(value)) return mergePluginConfig(value);
    return value;
}

function mergeInto(target: Record<string, unknown>, source: Record<string, unknown>) {
    for (const [key, value] of Object.entries(source)) {
        const current = Object.hasOwn(target, key) ? target[key] : undefined;
        if (key !== "categoryMapping" && isObject(value) && isObject(current)) {
            mergeInto(current, value);
        } else {
            Object.defineProperty(target, key, {
                value: clone(value), enumerable: true, writable: true, configurable: true,
            });
        }
    }
}
