import { z, ZodType } from "zod";
import type { Integration } from "../integration";

const StravaSchema = z.object({
    clientId: z.number({ coerce: true }).int().positive(),
    clientSecret: z.string().length(40),
    routes: z.boolean(),
    activities: z.boolean(),
    accessToken: z.string().length(40).optional(),
    refreshToken: z.string().length(40).optional(),
    expiresAt: z.number().int().positive().optional(),
    active: z.boolean()
})

const KomootSchema = z.object({
    email: z.string().email(),
    password: z.string(),
    active: z.boolean()
})

const IntegrationCreateSchema = z.object({
    user: z.string().length(15),
    strava: StravaSchema.optional(),
    komoot: KomootSchema.optional()

}) satisfies ZodType<Integration>

const IntegrationUpdateSchema = z.object({
    strava: StravaSchema.optional().nullable(),
    komoot: KomootSchema.optional().nullable()
}) satisfies ZodType<Partial<Integration>>

    export { StravaSchema, IntegrationCreateSchema, IntegrationUpdateSchema, KomootSchema };
