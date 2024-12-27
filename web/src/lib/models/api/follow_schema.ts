import { z, ZodType } from "zod";
import type { Follow } from "../follow";

const FollowCreateSchema = z.object({
    follower: z.string().length(15),
    followee: z.string().length(15),

}) satisfies ZodType<Follow>


export { FollowCreateSchema };
