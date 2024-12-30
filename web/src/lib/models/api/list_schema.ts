import { z, ZodType } from "zod";
import type { List } from "../list";

const ListCreateSchema = z.object({
    id: z.string().length(15).optional(),
    name: z.string().min(1, "required"),
    public: z.boolean(),
    description: z.string().optional(),
    trails: z.array(z.string().length(15)),
    author: z.string()
}) satisfies ZodType<List>

const ListUpdateSchema = z.object({
    name: z.string().optional(),
    public: z.boolean().optional(),
    description: z.string().optional(),
    trails: z.array(z.string().length(15)).optional(),
    "trails-": z.string().optional(),
    "trails+": z.string().optional(),
}) satisfies ZodType<Partial<List>>

export { ListCreateSchema, ListUpdateSchema };
