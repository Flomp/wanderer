import { z, ZodType } from "zod";
import type { Trail } from "../trail";


const TrailCreateSchema = z.object({
    name: z.string(),
    description: z.string().optional(),
    location: z.string().optional(),
    date: z.string().date(),
    public: z.boolean(),
    difficulty: z.enum(["easy", "moderate", "difficult"]).optional(),
    lat: z.number().optional(),
    lon: z.number().optional(),
    distance: z.number({coerce: true}).nonnegative().optional(),
    elevation_gain: z.number({coerce: true}).nonnegative().optional(),
    elevation_loss: z.number({coerce: true}).nonnegative().optional(),
    duration: z.number({coerce: true}).nonnegative().optional(),
    photos: z.array(z.string()),
    thumbnail: z.number().int().nonnegative().optional(),
    waypoints: z.array(z.string()),
    summit_logs: z.array(z.string()),
    category: z.string().optional(),
    gpx: z.string().optional(),
    author: z.string().length(15),

}) satisfies ZodType<Trail>

const TrailUpdateSchema = z.object({
    name: z.string(),
    description: z.string().optional(),
    location: z.string().optional(),
    date: z.string().date(),
    public: z.boolean().optional(),
    difficulty: z.enum(["easy", "moderate", "difficult"]).optional(),
    lat: z.number().optional(),
    lon: z.number().optional(),
    distance: z.number({coerce: true}).nonnegative().optional(),
    elevation_gain: z.number({coerce: true}).nonnegative().optional(),
    elevation_loss: z.number({coerce: true}).nonnegative().optional(),
    duration: z.number({coerce: true}).nonnegative().optional(),
    photos: z.array(z.string()).optional(),
    thumbnail: z.number().int().nonnegative().optional(),
    waypoints: z.array(z.string()).optional(),
    summit_logs: z.array(z.string()).optional(),
    category: z.string().optional(),
    gpx: z.string().optional(),
}) satisfies ZodType<Partial<Trail>>

export { TrailCreateSchema, TrailUpdateSchema };
