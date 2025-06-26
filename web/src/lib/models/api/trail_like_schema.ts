import { z, ZodType } from "zod";
import type { TrailLike } from "../trail_like";

const TrailLikeCreateSchema = z.object({
    actor: z.string().length(15),
    trail: z.string().length(15),

}) satisfies ZodType<TrailLike>

export { TrailLikeCreateSchema };
