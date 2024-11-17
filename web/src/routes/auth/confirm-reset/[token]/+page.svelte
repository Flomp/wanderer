<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import Button from "$lib/components/base/button.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import LogoTextTwoLineDark from "$lib/components/logo/logo_text_two_line_dark.svelte";
    import LogoTextTwoLineLight from "$lib/components/logo/logo_text_two_line_light.svelte";
    import type { User } from "$lib/models/user";
    import { theme } from "$lib/stores/theme_store";
    import { show_toast } from "$lib/stores/toast_store";
    import {
        users_confirm_reset,
        users_reset_password,
    } from "$lib/stores/user_store";
    import { createForm } from "$lib/vendor/svelte-form-lib";
    import { _ } from "svelte-i18n";
    import { object, ref, string } from "yup";

    let loading: boolean = false;

    const { form, errors, handleChange, handleSubmit } = createForm({
        initialValues: {
            password: "",
            passwordConfirm: "",
        },
        validationSchema: object({
            password: string().required($_("required")),
            passwordConfirm: string()
                .required($_("required"))
                .oneOf([ref("password"), ""], $_('passwords-must-match')),
        }),
        onSubmit: async (form) => {
            loading = true;
            try {
                await users_confirm_reset({
                    ...form,
                    token: $page.params.token,
                });
                goto("/login");
                show_toast({
                    icon: "check",
                    type: "success",
                    text: $_("new-password-success"),
                });
            } catch (e) {
                show_toast({
                    icon: "close",
                    type: "error",
                    text: $_("new-password-error"),
                });
            } finally {
                loading = false;
            }
        },
    });
</script>

<svelte:head>
    <title>{$_("login")} | wanderer</title>
</svelte:head>
<main class="flex justify-center">
    <form
        class="login-panel max-w-md border border-input-border rounded-xl p-8 flex flex-col justify-center items-center gap-4 w-[28rem] mt-8"
        on:submit={handleSubmit}
    >
        {#if $theme == "light"}
            <LogoTextTwoLineDark></LogoTextTwoLineDark>
        {:else}
            <LogoTextTwoLineLight></LogoTextTwoLineLight>
        {/if}
        <h4 class="text-xl font-semibold">{$_("new-password-text")}</h4>
        <div class="space-y-6 w-80">
            <TextField
                name="password"
                label={$_("password")}
                type="password"
                bind:value={$form.password}
                on:change={handleChange}
                error={$errors.password}
            ></TextField>
            <TextField
                name="password"
                label={$_("password-confirm")}
                type="password"
                bind:value={$form.passwordConfirm}
                on:change={handleChange}
                error={$errors.passwordConfirm}
            ></TextField>

            <Button
                primary={true}
                extraClasses={"min-w-full"}
                type="submit"
                {loading}>{$_("reset-password")}</Button
            >
        </div>
    </form>
</main>
