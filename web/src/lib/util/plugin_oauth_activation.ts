import type { PluginInstance } from "$lib/models/plugin_instance";

type Revision = Pick<PluginInstance, "id" | "connection_generation" | "updated">;
type IntentStorage = Pick<Storage, "getItem" | "setItem" | "removeItem">;
const prefix = "wanderer_plugin_repair_activation_";

function tabStorage(): IntentStorage | undefined {
    try { return typeof sessionStorage === "undefined" ? undefined : sessionStorage; }
    catch { return undefined; }
}

// Repair intent belongs to the exact result of our own write. In particular,
// a later explicit false->false disable must invalidate it even at the same G.
export function pluginOAuthActivationDefault(instance?: PluginInstance, storage = tabStorage()): boolean {
    if (!instance) return true;
    if (instance.enabled) {
        clearPluginOAuthRepairActivation(instance.id, storage);
        return true;
    }
    if (!instance.id || !storage) return false;
    try {
        const saved = JSON.parse(storage.getItem(prefix + instance.id) ?? "null");
        if (saved && instance.connection_generation && instance.updated &&
            saved.generation === instance.connection_generation && saved.updated === instance.updated &&
            typeof saved.activate === "boolean") return saved.activate;
        storage.removeItem(prefix + instance.id);
    } catch { /* Unavailable or invalid tab storage does not change the default. */ }
    return false;
}

export function rememberPluginOAuthRepairActivation(revision: Revision, activate: boolean, storage = tabStorage()) {
    if (!revision.id || !storage) return;
    try {
        if (!revision.connection_generation || !revision.updated) {
            storage.removeItem(prefix + revision.id);
            return;
        }
        storage.setItem(prefix + revision.id, JSON.stringify({
            generation: revision.connection_generation, updated: revision.updated, activate,
        }));
    } catch { /* Repair remains possible when tab storage is unavailable. */ }
}

export function clearPluginOAuthRepairActivation(id?: string, storage = tabStorage()) {
    if (!id || !storage) return;
    try { storage.removeItem(prefix + id); } catch { /* Optional tab memory. */ }
}
