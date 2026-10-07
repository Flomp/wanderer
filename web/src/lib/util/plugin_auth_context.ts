import type { PluginInstance } from "$lib/models/plugin_instance";
import type { PluginAuthContext, PluginProvider } from "$lib/models/plugin_provider";

export interface ResolvedPluginAuthContext {
    name?: string;
    auth: PluginAuthContext;
}

// Context identity is explicit, or is inferred only for a single-context manifest.
// Sorting multiple contexts would be deterministic but would still choose the wrong account.
export function resolvePluginAuthContext(
    contexts: Record<string, PluginAuthContext>,
    requested?: string,
): ResolvedPluginAuthContext {
    const names = Object.keys(contexts);
    if (requested) {
        if (!Object.hasOwn(contexts, requested)) {
            throw new Error("Unknown plugin authentication context");
        }
        return { name: requested, auth: contexts[requested] };
    }
    if (names.length === 0) return { auth: { type: "none" } };
    if (names.length !== 1) throw new Error("Plugin authentication context is ambiguous");
    const name = names[0];
    return { name, auth: contexts[name] };
}

export function resolveProviderAuth(
    plugin: PluginProvider,
    instance?: Partial<PluginInstance>,
): ResolvedPluginAuthContext {
    // Already normalized providers and older component callers remain supported.
    if (!plugin.authContexts) return { name: plugin.authContext, auth: plugin.auth };
    return resolvePluginAuthContext(plugin.authContexts, instance?.auth?.oauthContext);
}
