import { z, ZodType } from "zod";
import type { Follow } from "../follow";

const FollowCreateSchema = z.object({
    followee: z.string().length(15),

})


export { FollowCreateSchema };
