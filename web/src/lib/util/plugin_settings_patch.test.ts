import { describe, expect, it } from "vitest";
import { pluginSettingsConfigPatch, hasProviderConfigChanges } from "./plugin_settings_patch";

describe("intentional plugin config patches", () => {
    const opening = { plugin: { selector: "saved", baseURL: "" },
        host: { merge: { enabled: false }, connectors: { service: { baseURL: "" } } } };
    it("does not echo invalid stored configuration or defaults on unchanged saves", () => {
        expect(pluginSettingsConfigPatch(opening, structuredClone(opening))).toEqual({});
        expect(hasProviderConfigChanges({})).toBe(false);
    });
    it("patches one changed setting while preserving concurrent sibling values", () => {
        expect(pluginSettingsConfigPatch(opening, { ...opening, plugin: { ...opening.plugin, selector: "draft" } }))
            .toEqual({ plugin: { selector: "draft" } });
    });
    it("keeps deletion and category mapping replacement intentional", () => {
        expect(pluginSettingsConfigPatch({ plugin: { after: "2025-01-01" }, host: { categoryMapping: { hike: "old" } } },
            { plugin: {}, host: { categoryMapping: {} } }))
            .toEqual({ plugin: { after: null }, host: { categoryMapping: {} } });
    });
    it("exempts only the declared dismiss metadata from full config validation", () => {
        expect(hasProviderConfigChanges({ host: { categoryRemapDismissedAt: "date" } })).toBe(false);
        expect(hasProviderConfigChanges({ host: { categoryRemapDismissedAt: "date", connectors: {} } })).toBe(true);
    });
});
