import type { PluginInstance, PluginInstanceDraft, PluginInstanceSettings } from "$lib/models/plugin_instance";
import { APIError } from "$lib/util/api_util";
import { clearPluginOAuthRepairActivation } from "$lib/util/plugin_oauth_activation";
import type { ListResult } from "pocketbase";
import { writable, type Writable } from "svelte/store";

export const pluginInstances: Writable<PluginInstance[]> = writable([]);

export function rememberPluginInstance(instance: PluginInstance) {
    pluginInstances.update(instances => [...instances.filter(item => item.id !== instance.id), instance]);
    return instance;
}

export async function plugin_instances_show(
    id: string,
    f: typeof fetch = fetch,
    remember = true,
) {
    const r = await f(`/api/v1/plugin-instance/${id}`);
    const response = await r.json();
    if (!r.ok) throw new APIError(r.status, response.message, response);
    return remember ? rememberPluginInstance(response as PluginInstance) : response as PluginInstance;
}

function settingsBody(instance: PluginInstanceSettings): PluginInstanceSettings {
    const keys = ["enabled", "auth", "authContext", "config", "expected_connection_generation"] as const;
    return Object.fromEntries(keys.filter(key => instance[key] !== undefined).map(key => [key, instance[key]]));
}

export async function plugin_instances_index(
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
) {
    const r = await f("/api/v1/plugin-instance?perPage=-1", {
        method: "GET",
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response);
    }

    const fetchedInstances: ListResult<PluginInstance> = await r.json();
    pluginInstances.set(fetchedInstances.items);

    return fetchedInstances.items;
}

export async function plugin_instances_create(
    instance: PluginInstanceDraft,
    f: typeof fetch = fetch,
) {
    const body = settingsBody(instance);
    delete body.expected_connection_generation;
    const r = await f("/api/v1/plugin-instance", {
        method: "PUT",
        body: JSON.stringify({
            ...body,
            plugin_id: instance.plugin_id,
        }),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response);
    }

    return rememberPluginInstance((await r.json()) as PluginInstance);
}

export async function plugin_instances_update(
    instance: PluginInstanceSettings & { id: string },
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
) {
    if (!instance.id) {
        throw Error("Plugin instance has no id");
    }

    const r = await f("/api/v1/plugin-instance/" + instance.id, {
        method: "POST",
        body: JSON.stringify(settingsBody(instance)),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response);
    }

    if (instance.enabled !== undefined) clearPluginOAuthRepairActivation(instance.id);
    return rememberPluginInstance((await r.json()) as PluginInstance);
}

export async function plugin_oauth_start(
    data: {
        pluginId: string;
        instanceId?: string;
        authContext?: string;
        redirectUri: string;
        expected_connection_generation: string;
        activate_on_success: boolean;
    },
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
) {
    const r = await f("/api/v1/plugin-system/oauth/start", {
        method: "POST",
        body: JSON.stringify(data),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response);
    }

    return (await r.json()) as { url: string; state: string; instanceId: string; connection_generation: string; updated: string };
}

export async function plugin_auth_validate(
    data: {
        pluginId: string;
        instanceId?: string;
        authContext?: string;
        auth: Record<string, string | null>;
        config?: Record<string, unknown>;
        expected_connection_generation?: string;
    },
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
) {
    const r = await f("/api/v1/plugin-system/auth/validate", {
        method: "POST",
        body: JSON.stringify(data),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response);
    }

    return (await r.json()) as { ok: boolean; authContext?: string };
}

export async function plugin_oauth_callback(
    data: { instanceId: string; code: string; state: string },
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
) {
    const r = await f("/api/v1/plugin-system/oauth/callback", {
        method: "POST",
        body: JSON.stringify(data),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response);
    }

    clearPluginOAuthRepairActivation(data.instanceId);
    return (await r.json()) as { ok: boolean };
}

export async function plugin_oauth_revoke(
    instanceId: string,
    expected_connection_generation: string,
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
) {
    const r = await f("/api/v1/plugin-system/oauth/revoke", {
        method: "POST",
        body: JSON.stringify({ instanceId, expected_connection_generation }),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response);
    }

    const response = (await r.json()) as { ok: boolean; connection_generation: string; updated: string };
    clearPluginOAuthRepairActivation(instanceId);
    await plugin_instances_show(instanceId, f);
    return response;
}

export async function plugin_category_remap_preview(
    instanceId: string,
    config?: Record<string, unknown>,
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
) {
    const r = await f("/api/v1/plugin-system/category-remap/preview", {
        method: "POST",
        body: JSON.stringify({ instanceId, config }),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response);
    }

    return (await r.json()) as { count: number; backfilledSinceMapping?: number };
}

export async function plugin_category_remap_apply(
    instanceId: string,
    f: (url: RequestInfo | URL, config?: RequestInit) => Promise<Response> = fetch,
) {
    const r = await f("/api/v1/plugin-system/category-remap/apply", {
        method: "POST",
        body: JSON.stringify({ instanceId }),
    });

    if (!r.ok) {
        const response = await r.json();
        throw new APIError(r.status, response.message, response);
    }

    return (await r.json()) as { count: number; remapped?: number };
}
