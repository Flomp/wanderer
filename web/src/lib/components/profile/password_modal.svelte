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
        $form.password = ""
        $errors.password = "";
        $form.oldPassword = ""
        $errors.oldPassword = "";
        _openModal!();
    }

    const dispatch = createEventDispatcher();

    const { form, errors, handleChange, handleSubmit } = createForm<{
        oldPassword: string;
        password: string;
    }>({
        initialValues: { oldPassword: "", password: "" },
        validationSchema: object({
            oldPassword: string()
                .min(
                    8,
                    $_("must-be-at-least-n-characters-long", {
                        values: { n: 8 },
                    }),
                )
                .required($_("required")),
            password: string()
                .min(
                    8,
                    $_("must-be-at-least-n-characters-long", {
                        values: { n: 8 },
                    }),
                )
                .required($_("required")),
        }),
        onSubmit: async (_) => {
            dispatch("save", {oldPassword: $form.oldPassword, password: $form.password, passwordConfirm: $form.password});
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
    <form id="password-form" slot="content" on:submit={handleSubmit}>
        <TextField
            name="oldPassword"
            label={$_("current-password")}
            type="password"
            bind:value={$form.oldPassword}
            error={$errors.oldPassword}
            on:change={handleChange}
        ></TextField>
        <TextField
            name="password"
            label={$_("new-password")}
            type="password"
            bind:value={$form.password}
            error={$errors.password}
            on:change={handleChange}
        ></TextField>
    </form>
    <div slot="footer" class="flex items-center gap-4">
        <button class="btn-secondary" on:click={closeModal}
            >{$_("cancel")}</button
        >
        <button class="btn-primary" type="submit" form="password-form" name="save"
            >{$_("save")}</button
        >
    </div></Modal
>
