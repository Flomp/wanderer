<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";
    import TextField from "../base/text_field.svelte";
    import { validator } from "@felte/validator-zod";
    import { createForm } from "felte";
    import { z } from "zod";

    let _openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;
    export function openModal() {
        setFields("password", "")
        setFields("oldPassword", "")
        setErrors("password", [])
        setErrors("oldPassword", [])
        _openModal!();
    }

    const dispatch = createEventDispatcher();

    const { form, errors, setFields, setErrors } = createForm<{
        oldPassword: string;
        password: string;
        passwordConfirm: string;
    }>({
        initialValues: { oldPassword: "", password: "" },
        extend: validator({
            schema: z
                .object({
                    oldPassword: z.string().min(
                        8,
                        $_("must-be-at-least-n-characters-long", {
                            values: { n: 8 },
                        }),
                    ),
                    password: z.string().min(
                        8,
                        $_("must-be-at-least-n-characters-long", {
                            values: { n: 8 },
                        }),
                    ),
                    passwordConfirm: z.string().min(
                        8,
                        $_("must-be-at-least-n-characters-long", {
                            values: { n: 8 },
                        }),
                    ),
                })
                .refine((d) => d.password === d.passwordConfirm, {
                    message: "passwords-must-match",
                    path: ["passwordConfirm"],
                }),
        }),
        onSubmit: async (form) => {
            dispatch("save", {
                oldPassword: form.oldPassword,
                password: form.password,
                passwordConfirm: form.password,
            });
            closeModal!();
        },
    });
</script>

<Modal
    id="password-modal"
    size="max-w-sm"
    title={$_("change-password")}
    bind:openModal={_openModal}
    bind:closeModal
>
    <form id="password-form" slot="content" use:form>
        <TextField
            name="oldPassword"
            label={$_("current-password")}
            type="password"
            error={$errors.oldPassword}
        ></TextField>
        <TextField
            name="password"
            label={$_("new-password")}
            type="password"
            error={$errors.password}
        ></TextField>
        <TextField
            name="passwordConfirm"
            label={$_("password-confirm")}
            type="password"
            error={$errors.passwordConfirm}
        ></TextField>
    </form>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}
            >{$_("cancel")}</button
        >
        <button
            class="btn-primary"
            type="submit"
            form="password-form"
            name="save">{$_("save")}</button
        >
    </div></Modal
>
