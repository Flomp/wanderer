import { z, ZodType } from "zod";
import type { Waypoint } from "../waypoint";

const WaypointCreateSchema = z.object({
    id: z.string().length(15).optional(),
    name: z.string().optional(),
    description: z.string().optional(),
    lat: z.number({coerce: true}).min(-90).max(90),
    lon: z.number({coerce: true}).min(-180).max(180),
    icon: z.string().optional(),
    author: z.string().length(15),
    photos: z.array(z.string()).default([]),
}) satisfies ZodType<Partial<Waypoint>>

const WaypointUpdateSchema = z.object({
    name: z.string().optional(),
    description: z.string().optional(),
    lat: z.number({coerce: true}).min(-90).max(90).optional(),
    lon: z.number({coerce: true}).min(-180).max(180).optional(),
    icon: z.string().default("circle").optional(),
    photos: z.array(z.string()).optional(),
    "photos-": z.string().optional(),
    "photos+": z.string().optional(),
}) satisfies ZodType<Partial<Waypoint>>

export { WaypointCreateSchema, WaypointUpdateSchema };
