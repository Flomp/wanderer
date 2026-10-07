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
    if (!plugin.authContexts) {
        if (instance && plugin.auth.type !== "none" && !instance.auth_context) {
            throw new Error("Plugin authentication context is missing");
        }
        if (instance?.auth_context && plugin.authContext && instance.auth_context !== plugin.authContext) {
            throw new Error("Unknown plugin authentication context");
        }
        return { name: instance?.auth_context ?? plugin.authContext, auth: plugin.auth };
    }
    if (instance && Object.keys(plugin.authContexts).length > 0 && !instance.auth_context) {
        throw new Error("Plugin authentication context is missing");
    }
    return resolvePluginAuthContext(plugin.authContexts, instance?.auth_context);
}

export function editableAuthFields(auth: PluginAuthContext): string[] {
    const fields = new Set(auth.fields ?? []);
    if (auth.type === "oauth2") {
        fields.add("clientSecret");
    }
    if (auth.secretField) fields.add(auth.secretField);
    // Encryption declarations alone do not expose editable OAuth internals.
    const internal = new Set(["refreshToken", "expiresAt", "tokenType", "scope", "oauthContext",
        "providerUserId", "oauthState", "oauthCodeVerifier", "oauthRedirectURI",
        "oauthFlowContext", "oauthActivateOnCallback", "oauthAuthorizationId"]);
    if (auth.type === "oauth2") internal.add("accessToken");
    return [...fields].filter(field => !internal.has(field));
}

export function pluginSecretFields(plugin: PluginProvider): Set<string> {
    const fields = new Set(["accessToken", "refreshToken", "clientSecret"]);
    for (const auth of Object.values(plugin.authContexts ?? { active: plugin.auth })) {
        for (const field of auth.secretFields ?? []) fields.add(field);
        if (auth.secretField) fields.add(auth.secretField);
    }
    return fields;
}

export function hasPluginConnectionMaterial(plugin: PluginProvider, instance?: PluginInstance): boolean {
    if (!instance) return false;
    try {
        const { auth, name } = resolveProviderAuth(plugin, instance);
        if (auth.type === "none") return true;
        if (auth.type === "oauth2") {
            const provenance = instance.auth?.oauthContext;
            if (provenance !== undefined && provenance !== name) return false;
            if (provenance === undefined && Object.keys(plugin.authContexts ?? {}).length > 1) return false;
            return ["accessToken", "refreshToken"].some(key => Object.hasOwn(instance.auth ?? {}, key));
        }
        const fields = editableAuthFields(auth);
        const secrets = pluginSecretFields(plugin);
        return fields.length > 0 && fields.every(key => secrets.has(key)
            ? Object.hasOwn(instance.auth ?? {}, key) : !!instance.auth?.[key]);
    } catch {
        return false;
    }
}

export function canUsePluginConnection(plugin: PluginProvider, instance?: PluginInstance): boolean {
    return !!instance?.enabled && hasPluginConnectionMaterial(plugin, instance) &&
        instance.status !== "needs_auth" && instance.status !== "needs_reauth" &&
        instance.last_error?.code !== "credentials_unreadable";
}
