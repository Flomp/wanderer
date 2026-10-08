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
            if (provenance !== undefined && provenance !== "" && provenance !== name) return false;
            if ((provenance === undefined || provenance === "") && Object.keys(plugin.authContexts ?? {}).length > 1) return false;
            return ["accessToken", "refreshToken"].some(key => Object.hasOwn(instance.auth ?? {}, key));
        }
        const credentials = pluginAuthDraftInstance(plugin, instance, name)?.auth ?? {};
        const fields = editableAuthFields(auth);
        const secrets = pluginSecretFields(plugin);
        return fields.length > 0 && fields.every(key => secrets.has(key)
            ? Object.hasOwn(credentials, key) : !!credentials[key]);
    } catch {
        return false;
    }
}

export function hasStoredOAuthGrant(plugin: PluginProvider, instance?: PluginInstance): boolean {
    if (!instance || !["accessToken", "refreshToken"].some(key => Object.hasOwn(instance.auth ?? {}, key))) return false;
    // Redacted token keys preserve the host's evidence even when its former
    // context is no longer declared by the current manifest.
    if (Object.hasOwn(instance.auth ?? {}, "refreshToken") ||
        ["expiresAt", "tokenType", "oauthContext", "scope", "providerUserId"].some(key => Object.hasOwn(instance.auth ?? {}, key))) return true;
    try { return resolveProviderAuth(plugin, instance).auth.type === "oauth2"; }
    catch { return false; }
}

// A method preview cannot repurpose a saved OAuth grant as an API-key input.
// The host consumes that grant only after an explicit settings method change.
export function pluginAuthDraftInstance(plugin: PluginProvider, instance?: PluginInstance, requested?: string): PluginInstance | undefined {
    if (!instance) return undefined;
    const preview = { ...instance, auth_context: requested ?? instance.auth_context, auth: { ...instance.auth } };
    const provenance = instance.auth?.oauthContext;
    const selectionChangesUnattributedMaterial = !!instance.auth_context && !!requested && requested !== instance.auth_context &&
        (provenance === undefined || provenance === "");
    try {
        const selectingOAuth = resolveProviderAuth(plugin, preview).auth.type === "oauth2";
        if (selectingOAuth && selectionChangesUnattributedMaterial ||
            !selectingOAuth && hasStoredOAuthGrant(plugin, instance)) {
            for (const key of ["accessToken", "refreshToken", "expiresAt", "tokenType", "scope", "oauthContext", "providerUserId"]) {
                delete preview.auth[key];
            }
        }
    } catch { /* An unresolved preview has no usable connection material. */ }
    return preview;
}

export function canUsePluginConnection(plugin: PluginProvider, instance?: PluginInstance): boolean {
    return !!instance?.enabled && hasPluginConnectionMaterial(plugin, instance) &&
        instance.status !== "needs_auth" && instance.status !== "needs_reauth" &&
        instance.last_error?.code !== "credentials_unreadable";
}
