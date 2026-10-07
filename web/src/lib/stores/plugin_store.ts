import type { LocalizedTextMap, PluginAuthContext, PluginProvider } from "$lib/models/plugin_provider";
import type { PluginSystemPlugin } from "$lib/models/plugin_system";
import { APIError } from "$lib/util/api_util";
import { resolvePluginAuthContext, canUsePluginConnection } from "$lib/util/plugin_auth_context";
import { derived, writable, type Readable, type Writable } from "svelte/store";
import {
    pluginInstances,
    plugin_instances_index,
} from "./plugin_instance_store";

export const pluginProviders: Writable<PluginProvider[]> = writable([]);

// True when the user has at least one enabled plugin instance whose plugin
// advertises the trail send capability. Used to gate the trail "send to" action.
export const hasSendCapablePlugin: Readable<boolean> = derived(
    [pluginProviders, pluginInstances],
    ([$plugins, $instances]) => {
        return $plugins.some(plugin => plugin.status === "available" &&
            (plugin.capabilities ?? []).includes("prepare_trail_send.v1") &&
            $instances.some(instance => instance.plugin_id === plugin.id && canUsePluginConnection(plugin, instance)));

    },
);

let pluginDataLoaded = false;

// Loads plugins and instances once per session so the derived gating
// stores are populated (e.g. for the trail "send to" action). Safe to call
// repeatedly; only the first call performs the requests.
export async function load_plugin_data_once() {
    if (pluginDataLoaded) {
        return;
    }
    pluginDataLoaded = true;
    try {
        await Promise.all([
            plugins_index(),
            plugin_instances_index(),
        ]);
    } catch (e) {
        pluginDataLoaded = false;
        console.error(e);
    }
}

export async function plugins_index(
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
) {
    const r = await f("/api/v1/plugin-system/plugins", {
        method: "GET",
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response.detail);
    }

    const data: { items: PluginSystemPlugin[] } = await r.json();
    const items = data.items.map(pluginSystemToPluginProvider);
    pluginProviders.set(items);

    return items;
}

function pluginSystemToPluginProvider(plugin: PluginSystemPlugin): PluginProvider {
    const contexts = (plugin.manifest.auth?.contexts ?? {}) as Record<string, PluginAuthContext>;
    let primaryAuth: PluginAuthContext = { type: "unresolved" };
    let authContext: string | undefined;
    let authContextError: string | undefined;
    try {
        const resolved = resolvePluginAuthContext(contexts);
        primaryAuth = resolved.auth;
        authContext = resolved.name;
    } catch {
        authContextError = "auth_context_ambiguous";
        // An existing instance may have an explicit context. Defer resolution
        // until it is available; never pick a context from manifest map order.
    }
    const fields =
        primaryAuth.fields ??
        primaryAuth.secretFields ??
        (primaryAuth.secretField ? [primaryAuth.secretField] : []);
    const metadata = plugin.manifest.metadata ?? {};
    const setupErrorCode =
        typeof plugin.setupErrorCode === "string"
            ? plugin.setupErrorCode.trim() || undefined
            : undefined;

    return {
        id: plugin.id,
        type: plugin.type ?? plugin.manifest.type ?? "trails",
        name: plugin.name,
        displayName: plugin.displayName,
        displayNames: localizedMetadata(metadata, "displayNames"),
        description: plugin.description,
        descriptions: localizedMetadata(metadata, "descriptions"),
        information: localizedMetadata(metadata, "information"),
        homepageUrl: externalHttpUrl(metadata.homepageUrl),
        donationUrl: externalHttpUrl(metadata.donationUrl),
        icon: plugin.icon,
        iconDark: plugin.iconDark,
        version: plugin.version,
        auth: {
            type: primaryAuth.type ?? "none",
            fields,
            secretFields: primaryAuth.secretFields ?? fields,
            authorizationUrl: primaryAuth.authorizationUrl,
            tokenUrl: primaryAuth.tokenUrl,
            tokenRequestFormat: primaryAuth.tokenRequestFormat,
            scopes: primaryAuth.scopes,
            scopeSeparator: primaryAuth.scopeSeparator,
            authorizationParams: primaryAuth.authorizationParams,
            pkce: primaryAuth.pkce,
            tokenAuth: primaryAuth.tokenAuth,
        },
        authContext,
        authContextError,
        authContexts: contexts,
        configSchema: plugin.manifest.configSchema as PluginProvider["configSchema"],
        hostConfig: plugin.manifest.hostConfig,
        metadata,
        capabilities: plugin.capabilities,
        status: plugin.status,
        setupErrorCode,
        error: plugin.error,
    };
}

// Exported so the security boundary for plugin-provided links can be tested directly.
export function externalHttpUrl(value: unknown): string | undefined {
    if (typeof value !== "string") {
        return undefined;
    }
    try {
        const url = new URL(value);
        return url.protocol === "https:" || url.protocol === "http:"
            ? url.toString()
            : undefined;
    } catch {
        return undefined;
    }
}

function localizedMetadata(
    metadata: Record<string, unknown>,
    key: string,
): LocalizedTextMap | undefined {
    const value = metadata[key];
    if (!value || typeof value !== "object" || Array.isArray(value)) {
        return undefined;
    }
    const entries = Object.entries(value as Record<string, unknown>).filter(
        (entry): entry is [string, string] => typeof entry[1] === "string",
    );
    return entries.length ? Object.fromEntries(entries) : undefined;
}
