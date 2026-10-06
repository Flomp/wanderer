import { describe, expect, it } from "vitest";
import { PluginSessionAuthValidateSchema } from "./plugin_instance_schema";

describe("session auth validation draft", () => {
    const data = { pluginId: "configured-session", instanceId: "abcdefghijklmn1",
        authContext: "session", auth: { password: "replacement", email: "test@example.org" },
        config: { host: { connectors: { service: { baseURL: "http://private.example", allowPrivate: true,
            tls: { mode: "customCA", caBundle: "draft-ca" } } } }, plugin: { selector: "unsaved" } } };

    it("keeps draft config and replacement credentials for the host without coercion", () => {
        expect(PluginSessionAuthValidateSchema.parse(data)).toEqual(data);
        expect(PluginSessionAuthValidateSchema.parse({ pluginId: "legacy", auth: {} }))
            .toEqual({ pluginId: "legacy", auth: {} });
    });

    it.each([null, [], "config"])("rejects a non-object config draft %j", config => {
        expect(PluginSessionAuthValidateSchema.safeParse({ ...data, config }).success).toBe(false);
    });

    it("rejects malformed context, identity and credential types", () => {
        for (const patch of [{ instanceId: "invalid" }, { pluginId: "../plugin" }, { authContext: "" },
            { auth: { password: null } }, { auth: [] }]) {
            expect(PluginSessionAuthValidateSchema.safeParse({ ...data, ...patch }).success).toBe(false);
        }
    });
});
