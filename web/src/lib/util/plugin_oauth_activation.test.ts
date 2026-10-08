import { describe, expect, it } from "vitest";
import { clearPluginOAuthRepairActivation, pluginOAuthActivationDefault, rememberPluginOAuthRepairActivation } from "./plugin_oauth_activation";
import type { PluginInstance } from "$lib/models/plugin_instance";

function fixture() {
    const values = new Map<string, string>();
    const storage = { getItem: (key: string) => values.get(key) ?? null,
        setItem: (key: string, value: string) => { values.set(key, value); },
        removeItem: (key: string) => { values.delete(key); } };
    const instance: PluginInstance = { id: "instance", plugin_id: "oauth", user: "owner", enabled: false,
        status: "disabled", connection_generation: "repair-generation", updated: "repair-commit" };
    return { storage, instance };
}

describe("OAuth repair activation choice", () => {
    it("defaults to active/new and preserves either chosen value only for our exact repair commit", () => {
        const { instance, storage } = fixture();
        expect(pluginOAuthActivationDefault(undefined, storage)).toBe(true);
        expect(pluginOAuthActivationDefault({ ...instance, enabled: true }, storage)).toBe(true);
        expect(pluginOAuthActivationDefault(instance, storage)).toBe(false);
        for (const activate of [true, false]) {
            rememberPluginOAuthRepairActivation(instance, activate, storage);
            expect(pluginOAuthActivationDefault(instance, storage)).toBe(activate);
        }
    });

    it("forgets intent after external revoke or same-generation explicit disable and cannot resurrect it", () => {
        const { instance, storage } = fixture();
        for (const external of [{ ...instance, connection_generation: "external-revoke" },
            { ...instance, updated: "external-false-to-false-disable" }]) {
            rememberPluginOAuthRepairActivation(instance, true, storage);
            expect(pluginOAuthActivationDefault(external, storage)).toBe(false);
            expect(pluginOAuthActivationDefault(instance, storage)).toBe(false);
        }
    });

    it("never transfers a choice to another instance or incomplete revision", () => {
        const { instance, storage } = fixture();
        rememberPluginOAuthRepairActivation(instance, true, storage);
        expect(pluginOAuthActivationDefault({ ...instance, id: "other" }, storage)).toBe(false);
        expect(pluginOAuthActivationDefault({ ...instance, updated: undefined }, storage)).toBe(false);
        rememberPluginOAuthRepairActivation({ ...instance, updated: undefined }, true, storage);
        expect(pluginOAuthActivationDefault(instance, storage)).toBe(false);
        rememberPluginOAuthRepairActivation(instance, true, storage);
        clearPluginOAuthRepairActivation(instance.id, storage);
        expect(pluginOAuthActivationDefault(instance, storage)).toBe(false);
    });
});
