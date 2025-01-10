import { z, ZodType } from "zod";
import { NotificationType } from "../notification";
import type { Settings } from "../settings";

const SettingsCreateSchema = z.object({
    unit: z.enum(["metric", "imperial"]).optional(),
    language: z.enum(["en", "de", "es", "fr", "hu", "it", "nl", "pl", "pt", "zh"]).optional(),
    mapFocus: z.enum(["trails", "location"]).optional(),
    location: z.object({
        name: z.string(),
        lat: z.number(),
        lon: z.number()
    }).optional(),
    category: z.string().optional(),
    tilesets: z.array(z.object({ name: z.string(), url: z.string().url() })).optional(),
    terrain: z.object({ terrain: z.string().url(), hillshading: z.string().url() }).optional(),
    user: z.string().optional(),
    privacy: z.object({
        account: z.enum(["public", "private"]),
        trails: z.enum(["public", "private"]),
        lists: z.enum(["public", "private"])
    }).optional(),
    notifications: z.record(z.enum(Object.values(NotificationType) as [string, ...string[]]), z.object({ web: z.boolean(), email: z.boolean() })).optional()

}) satisfies ZodType<Settings>
ZodType<Partial<Comment>>

export { SettingsCreateSchema };
