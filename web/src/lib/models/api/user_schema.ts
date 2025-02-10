import { z, ZodType } from "zod";
import type { User } from "../user";

const UserCreateSchema = (z.object({
    username: z.string({ message: "required" }).min(3, "must-be-at-least-n-characters-long").regex(/^[\w][\w\.]*$/, "invalid-username"),
    email: z.string({ message: "required" }).email("not-a-valid-email-address"),
    password: z.string().min(8, "must-be-at-least-n-characters-long"),
    passwordConfirm: z.string().optional()

}) satisfies ZodType<Partial<User>>).refine((data) => data.password === data.passwordConfirm, {
    message: "passwords-must-match",
    path: ["passwordConfirm"],
})

const UserUpdateSchema = (z.object({
    username: z.string({ message: "required" }).min(3, "must-be-at-least-n-characters-long").regex(/^[\w][\w\.]*$/, "invalid-username").optional(),
    email: z.string({ message: "required" }).email("not-a-valid-email-address").optional(),
    password: z.string().min(8, "must-be-at-least-n-characters-long").optional(),
    oldPassword: z.string().min(8, "must-be-at-least-n-characters-long").optional(),
    passwordConfirm: z.string().min(8, "must-be-at-least-n-characters-long").optional(),
}) satisfies ZodType<Partial<User>>).refine((data) => data.password === data.passwordConfirm, {
    message: "passwords-must-match",
    path: ["passwordConfirm"],
});



export { UserCreateSchema, UserUpdateSchema };
