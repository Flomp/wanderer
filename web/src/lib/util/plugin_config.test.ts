import { describe, expect, it } from "vitest";
import { mergePluginConfig } from "./plugin_config";

describe("plugin config layering", () => {
    it("preserves sibling connector defaults while overlaying draft fields", () => {
        const defaults = { connectors: { service: { baseURL: "https://installed.example", allowPrivate: true,
            tls: { mode: "customCA", caBundle: "existing-ca" } } }, merge: { enabled: false, available: true } };
        const overrides = { connectors: { service: { baseURL: "https://draft.example" } }, merge: { enabled: true } };
        expect(mergePluginConfig(defaults, overrides)).toEqual({
            connectors: { service: { baseURL: "https://draft.example", allowPrivate: true,
                tls: { mode: "customCA", caBundle: "existing-ca" } } },
            merge: { enabled: true, available: true },
        });
    });

    it("replaces category maps including deliberate empty maps", () => {
        expect(mergePluginConfig({ categoryMapping: { hiking: "default" } }, { categoryMapping: {} }))
            .toEqual({ categoryMapping: {} });
        expect(mergePluginConfig({ categoryMapping: { hiking: "default" } }, { categoryMapping: { cycling: "override" } }))
            .toEqual({ categoryMapping: { cycling: "override" } });
    });

    it("keeps JSON object keys without applying prototype setters", () => {
        const overrides = JSON.parse('{"__proto__":{"injected":true},"constructor":{"value":"config"}}');
        const result = mergePluginConfig(overrides);
        expect(Object.hasOwn(result, "__proto__")).toBe(true);
        expect(Object.getPrototypeOf(result)).toBe(Object.prototype);
        expect(result.constructor).toEqual({ value: "config" });
        expect(({} as Record<string, unknown>).injected).toBeUndefined();
    });

    it("makes a draft independent from caller defaults and nested arrays", () => {
        const defaults = { connectors: { service: { allowed: ["/one"] } } };
        const merged = mergePluginConfig(defaults) as typeof defaults;
        merged.connectors.service.allowed.push("/two");
        expect(defaults.connectors.service.allowed).toEqual(["/one"]);
    });
});
