<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import { validator } from "@felte/validator-zod";
    import { createForm } from "felte";
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";
    import { z } from "zod";
    import TextField from "../base/text_field.svelte";

    interface Props {
        email?: string;
    }

    let { email = "" }: Props = $props();

    let modal: Modal;

    export function openModal() {
        setFields("email", email);
        setErrors("email", []);
        modal.openModal();
    }

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
            modal.closeModal!();
        },
    });
</script>

<Modal
    id="email-modal"
    size="max-w-xl"
    title={$_("change-email")}
    bind:this={modal}
>
    {#snippet content()}
        <form id="email-form" use:form>
            <TextField name="email" error={$errors.email}></TextField>
        </form>
    {/snippet}
    {#snippet footer()}
        <div class="flex items-center gap-4">
            <button class="btn-secondary" onclick={() => modal.closeModal()}
                >{$_("cancel")}</button
            >
            <button
                class="btn-primary"
                type="submit"
                form="email-form"
                name="save">{$_("save")}</button
            >
        </div>
    {/snippet}</Modal
>
