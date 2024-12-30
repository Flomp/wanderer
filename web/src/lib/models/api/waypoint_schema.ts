import { z, ZodType } from "zod";
import type { Waypoint } from "../waypoint";

const WaypointCreateSchema = z.object({
    id: z.string().length(15).optional(),
    name: z.string().optional(),
    description: z.string().optional(),
    text: z.string().optional(),
    lat: z.number(),
    lon: z.number(),
    icon: z.string().optional(),
    author: z.string().length(15),
    photos: z.array(z.string()),
}) satisfies ZodType<Waypoint>

const WaypointUpdateSchema = z.object({
    name: z.string().optional(),
    description: z.string().optional(),
    text: z.string().optional(),
    lat: z.number({coerce: true}).optional(),
    lon: z.number({coerce: true}).optional(),
    icon: z.string().optional(),
    photos: z.array(z.string()),
    "photos-": z.string().optional(),
    "photos+": z.string().optional(),
}) satisfies ZodType<Partial<Waypoint>>

export { WaypointCreateSchema, WaypointUpdateSchema };
