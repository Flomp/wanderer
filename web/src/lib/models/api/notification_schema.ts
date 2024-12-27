import { z, type ZodType } from "zod"
import { type Notification } from "../notification";

const NotificationUpdateSchema = z.object({
    seen: z.literal(true)
}) satisfies ZodType<Partial<Notification>>

export { NotificationUpdateSchema }