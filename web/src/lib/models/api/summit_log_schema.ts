import { z, ZodType } from "zod";
import type { SummitLog } from "../summit_log";


const SummitLogCreateSchema = z.object({
    id: z.string().length(15).optional(),
    date: z.string().date().refine((val) => !isNaN(Date.parse(val)), "invalid-date"),
    text: z.string().optional(),
    gpx: z.string().optional(),
    distance: z.number().nonnegative().optional(),
    elevation_gain: z.number().nonnegative().optional(),
    elevation_loss: z.number().nonnegative().optional(),
    duration: z.number().nonnegative().optional(),
    author: z.string().length(15),
    photos: z.array(z.string()).default([])
}) satisfies ZodType<Partial<SummitLog>>

const SummitLogUpdateSchema = z.object({
    date: z.string().date().refine((val) => !val || !isNaN(Date.parse(val)), "invalid-date").optional(),
    text: z.string().optional(),
    gpx: z.string().optional(),
    distance: z.number().nonnegative().optional(),
    elevation_gain: z.number().nonnegative().optional(),
    elevation_loss: z.number().nonnegative().optional(),
    duration: z.number().nonnegative().optional(),
    photos: z.array(z.string()).optional(),
    "photos-": z.string().optional(),
    "photos+": z.string().optional(),
}) satisfies ZodType<Partial<SummitLog>>

export { SummitLogCreateSchema, SummitLogUpdateSchema };
