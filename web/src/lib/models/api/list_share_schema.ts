import { z, ZodType } from "zod";
import type { ListShare } from "../list_share";

const ListShareCreateSchema = z.object({
    actor: z.string().url(),
    list: z.string().length(15),
    permission: z.enum(["view", "edit"])

}) satisfies ZodType<ListShare>

const ListShareUpdateSchema = z.object({
    permission: z.enum(["view", "edit"])
}) satisfies ZodType<Partial<ListShare>>


export { ListShareCreateSchema, ListShareUpdateSchema };
