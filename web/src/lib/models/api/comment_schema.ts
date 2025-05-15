import { z, ZodType } from "zod";
import type { Comment } from "../comment";

const CommentCreateSchema = z.object({
    text: z.string(),
    author: z.string().length(15),
    trail: z.string(),
    handle: z.string().optional(),

}) satisfies ZodType<Comment>

const CommentUpdateSchema = z.object({
    text: z.string().optional(),
    handle: z.string().optional(),
}) satisfies ZodType<Partial<Comment>>

export { CommentCreateSchema, CommentUpdateSchema }