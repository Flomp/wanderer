<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";
    import TextField from "../base/text_field.svelte";
    import { createForm } from "$lib/vendor/svelte-form-lib";
    import { object, string } from "yup";

    let _openModal: (() => void) | undefined = undefined;
    export let closeModal: (() => void) | undefined = undefined;
    export function openModal() {
        $form.email = email
        $errors.email = "";
        _openModal!();
    }

    export let email = "";

    const dispatch = createEventDispatcher();

    const { form, errors, handleChange, handleSubmit } = createForm<{
        email: string;
    }>({
        initialValues: { email: email },
        validationSchema: object({
            email: string()
                .email($_("not-a-valid-email-address"))
                .required($_("required")),
        }),
        onSubmit: async (_) => {
            dispatch("save", $form.email);
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
    <form id="email-form" slot="content" on:submit={handleSubmit}>
        <TextField
            name="email"
            bind:value={$form.email}
            error={$errors.email}
            on:change={handleChange}
        ></TextField>
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
