import { z, ZodType } from "zod";
import type { Comment } from "../comment";

const CommentCreateSchema = z.object({
    text: z.string(),
    rating: z.number(),
    author: z.string().length(15),
    trail: z.string().length(15),

}) satisfies ZodType<Comment>

const CommentUpdateSchema = z.object({
    text: z.string().optional(),
    rating: z.number().optional(),
}) satisfies ZodType<Partial<Comment>>

export { CommentCreateSchema, CommentUpdateSchema }