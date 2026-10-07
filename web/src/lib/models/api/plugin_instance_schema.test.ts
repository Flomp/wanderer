import { describe, expect, it } from "vitest";
import { PluginSessionAuthValidateSchema, PluginInstanceCreateSchema, PluginInstanceUpdateSchema } from "./plugin_instance_schema";

describe("session auth validation draft", () => {
    const data = { pluginId: "configured-session", instanceId: "abcdefghijklmn1",
        authContext: "session", expected_connection_generation: "revision", auth: { password: "replacement", email: "test@example.org" },
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
            { auth: { password: 123 } }, { auth: [] }]) {
            expect(PluginSessionAuthValidateSchema.safeParse({ ...data, ...patch }).success).toBe(false);
        }
    });

    it("retains explicit credential clear and empty nonsecret input for validation", () => {
        const auth = { password: null, email: "" };
        expect(PluginSessionAuthValidateSchema.parse({ ...data, auth }).auth).toEqual(auth);
    });
});

describe("host instance write DTOs", () => {
    it("preserves omitted keys, empty secret placeholders and explicit null clears", () => {
        expect(PluginInstanceUpdateSchema.parse({})).toEqual({});
        const patch = { auth: { password: null, clientSecret: "", accessToken: "replacement" },
            authContext: "api_key", expected_connection_generation: "opaque-revision" };
        expect(PluginInstanceUpdateSchema.parse(patch)).toEqual(patch);
        expect(PluginInstanceCreateSchema.parse({ plugin_id: "test" })).toEqual({ plugin_id: "test" });
    });
    it.each(["user", "plugin_id", "status", "state", "connection_generation", "auth_context", "busy", "webhooksecret"])("rejects protected settings key %s", key => {
        expect(PluginInstanceUpdateSchema.safeParse({ [key]: "value" }).success).toBe(false);
    });
    it.each([null, [], "body"])("rejects non-object DTO %j", value => {
        expect(PluginInstanceCreateSchema.safeParse(value).success).toBe(false);
        expect(PluginInstanceUpdateSchema.safeParse(value).success).toBe(false);
    });
    it("does not coerce malformed auth values", () => {
        expect(PluginInstanceUpdateSchema.safeParse({ auth: { password: 123 } }).success).toBe(false);
        expect(PluginInstanceCreateSchema.safeParse({ plugin_id: "test", expected_connection_generation: "existing" }).success).toBe(false);
    });
});
