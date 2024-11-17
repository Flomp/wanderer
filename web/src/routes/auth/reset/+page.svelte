<script lang="ts">
    import { goto } from "$app/navigation";
    import Button from "$lib/components/base/button.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import LogoTextTwoLineDark from "$lib/components/logo/logo_text_two_line_dark.svelte";
    import LogoTextTwoLineLight from "$lib/components/logo/logo_text_two_line_light.svelte";
    import type { User } from "$lib/models/user";
    import { theme } from "$lib/stores/theme_store";
    import { show_toast } from "$lib/stores/toast_store";
    import { users_reset_password } from "$lib/stores/user_store";
    import { createForm } from "$lib/vendor/svelte-form-lib";
    import { _ } from "svelte-i18n";
    import { object, string } from "yup";

    let loading: boolean = false;

    const { form, errors, handleChange, handleSubmit } = createForm({
        initialValues: {
            email: "",
        },
        validationSchema: object<User>({
            email: string().required($_("required")),
        }),
        onSubmit: async (email) => {
            loading = true;
            try {
                await users_reset_password(email);
                goto("/");
                show_toast({
                    icon: "check",
                    type: "success",
                    text: $_("password-reset-sent"),
                });
            } catch (e) {
                show_toast({
                    icon: "close",
                    type: "error",
                    text: $_("error-during-password-reset"),
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
        <h4 class="text-xl font-semibold">{$_("forgot-your-password")}</h4>
        <p class="text-gray-500">{$_('password-reset-text')}</p>
        <div class="space-y-6 w-80">
            <TextField
                name="email"
                label={$_("email")}
                bind:value={$form.email}
                on:change={handleChange}
                error={$errors.email}
            ></TextField>
            <div class="grid grid-cols-2 gap-x-4">
                <a class="btn-secondary text-center" href="/login"
                    >{$_("back-to-login")}</a
                >
                <Button
                    primary={true}
                    extraClasses={"min-w-full"}
                    type="submit"
                    {loading}>{$_("reset-password")}</Button
                >
            </div>
        </div>
    </form>
</main>
