<script lang="ts">
    import { goto } from "$app/navigation";
    import Button from "$lib/components/base/button.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import LogoTextTwoLineDark from "$lib/components/logo/logo_text_two_line_dark.svelte";
    import LogoTextTwoLineLight from "$lib/components/logo/logo_text_two_line_light.svelte";
    import type { User } from "$lib/models/user";
    import { theme } from "$lib/stores/theme_store";
    import { show_toast } from "$lib/stores/toast_store";
    import { login, users_create } from "$lib/stores/user_store";
    import { createForm } from "$lib/vendor/svelte-form-lib";
    import { _ } from "svelte-i18n";
    import { object, string } from "yup";
    let loading: boolean = false;
    const { form, errors, handleChange, handleSubmit } = createForm<User>({
        initialValues: {
            id: "",
            username: "",
            email: "",
            password: "",
        },
        validationSchema: object<User>({
            username: string()
                .min(
                    3,
                    $_("must-be-at-least-n-characters-long", {
                        values: { n: 3 },
                    }),
                )
                .required($_("required"))
                .matches(/^[\w][\w\.]*$/, { message: $_("invalid-username") }),
            email: string()
                .email($_("not-a-valid-email-address"))
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
        onSubmit: async (newUser) => {
            loading = true;
            try {
                await users_create(newUser);
            } catch (e) {
                show_toast({
                    icon: "close",
                    type: "error",
                    text: $_("error-creating-user"),
                });
                return;
            } finally {
                loading = false;
            }
            try {
                await login(newUser);
                goto(`/`);
            } catch (e) {
                show_toast({
                    icon: "close",
                    type: "error",
                    text: $_("error-during-login"),
                });
            } finally {
                loading = false;
            }
        },
    });
</script>

<svelte:head>
    <title>{$_("register")} | wanderer</title>
</svelte:head>
<main class="flex justify-center">
    <form
        class="login-panel max-w-md border border-input-border rounded-xl p-8 flex flex-col justify-center items-center gap-8 w-[28rem] mt-8"
        on:submit={handleSubmit}
    >
        {#if $theme == "light"}
            <LogoTextTwoLineDark></LogoTextTwoLineDark>
        {:else}
            <LogoTextTwoLineLight></LogoTextTwoLineLight>
        {/if}
        <h4 class="text-xl font-semibold">{$_("slogan")}</h4>
        <div class="space-y-6 w-80">
            <TextField
                name="username"
                label={$_("username")}
                bind:value={$form.username}
                on:change={handleChange}
                error={$errors.username}
            ></TextField>
            <TextField
                name="email"
                label={$_("email")}
                bind:value={$form.email}
                on:change={handleChange}
                error={$errors.email}
            ></TextField>
            <TextField
                name="password"
                label={$_("password")}
                type="password"
                bind:value={$form.password}
                on:change={handleChange}
                error={$errors.password}
            ></TextField>
            <Button
                primary={true}
                extraClasses={"min-w-full"}
                type="submit"
                {loading}>{$_("register")}</Button
            >
        </div>
        <span
            >{$_("already-account")}
            <a class="text-blue-500 underline" href="/login">{$_("login")}!</a
            ></span
        >
    </form>
</main>
