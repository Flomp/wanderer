import { z } from "zod";

const pluginId = z.string().min(1).max(64).regex(/^[a-z0-9][a-z0-9_-]*$/);
const generation = z.string().min(1);
const settings = {
    enabled: z.boolean().optional(),
    auth: z.record(z.string(), z.string().nullable()).optional(),
    authContext: z.string().min(1).optional(),
    config: z.record(z.string(), z.unknown()).optional(),
    expected_connection_generation: generation.optional(),
};

export const PluginSessionAuthValidateSchema = z.object({
    pluginId,
    instanceId: z.string().length(15).optional(),
    authContext: z.string().min(1).optional(),
    auth: z.record(z.string(), z.string().nullable()),
    config: z.record(z.string(), z.unknown()).optional(),
    expected_connection_generation: generation.optional(),
}).strict();

export const PluginInstanceUpdateSchema = z.object(settings).strict();

export const PluginInstanceCreateSchema = PluginInstanceUpdateSchema
    .omit({ expected_connection_generation: true }).extend({ plugin_id: pluginId });
