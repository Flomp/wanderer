<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import { validator } from "@felte/validator-zod";
    import { createForm } from "felte";
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";
    import { z } from "zod";
    import TextField from "../base/text_field.svelte";

    let _openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;
    export function openModal() {
        setFields("email", email);
        setErrors("email", []);
        _openModal!();
    }

    export let email = "";

    const dispatch = createEventDispatcher();

    const { form, errors, setFields, setErrors } = createForm<{
        email: string;
    }>({
        initialValues: { email: email },
        extend: validator({
            schema: z.object({
                email: z
                    .string()
                    .min(1, "required")
                    .email("not-a-valid-email-address"),
            }),
        }),
        onSubmit: async (form) => {
            dispatch("save", form.email);
            closeModal!();
        },
    });
</script>

<Modal
    id="email-modal"
    size="max-w-xl"
    title={$_("change-email")}
    bind:openModal={_openModal}
    bind:closeModal
>
    <form id="email-form" slot="content" use:form>
        <TextField name="email" error={$errors.email}></TextField>
    </form>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}
            >{$_("cancel")}</button
        >
        <button class="btn-primary" type="submit" form="email-form" name="save"
            >{$_("save")}</button
        >
    </div></Modal
>
