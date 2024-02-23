<script lang="ts">
    import { goto } from "$app/navigation";
    import Button from "$lib/components/base/button.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import LogoTextTwoLineDark from "$lib/components/logo/logo_text_two_line_dark.svelte";
    import LogoTextTwoLineLight from "$lib/components/logo/logo_text_two_line_light.svelte";
    import { show_toast } from "$lib/stores/toast_store";
    import { login, type User } from "$lib/stores/user_store";
    import { createForm } from "$lib/vendor/svelte-form-lib";
    import { ClientResponseError } from "pocketbase";
    import { object, string } from "yup";
    import { theme } from "$lib/stores/theme_store";


    let loading: boolean = false;
    const { form, errors, handleChange, handleSubmit } = createForm<User>({
        initialValues: {
            username: "",
            password: "",
        },
        validationSchema: object<User>({
            username: string().required("Required"),
            password: string().required("Required"),
        }),
        onSubmit: async (newUser) => {
            loading = true;
            try {
                await login(newUser);
                goto(`/`);
            } catch (e) {
                if (
                    e instanceof ClientResponseError &&
                    e.message == "Failed to authenticate."
                ) {
                    show_toast({
                        icon: "close",
                        type: "error",
                        text: "Wrong username or password.",
                    });
                } else {
                    show_toast({
                        icon: "close",
                        type: "error",
                        text: "Error during login.",
                    });
                }
            } finally {
                loading = false;
            }
        },
    });
</script>

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
        <h4 class="text-xl font-semibold">Save your adventures!</h4>
        <div class="space-y-6 w-80">
            <TextField
                name="username"
                label="Username/Email"
                bind:value={$form.username}
                on:change={handleChange}
                error={$errors.username}
            ></TextField>
            <TextField
                name="password"
                label="Password"
                type="password"
                bind:value={$form.password}
                on:change={handleChange}
                error={$errors.password}
            ></TextField>
            <Button
                primary={true}
                extraClasses={"min-w-full"}
                type="submit"
                {loading}>Login</Button
            >
        </div>
        <span
            >Don't have an account? <a
                class="text-blue-500 underline"
                href="/register">Make one!</a
            ></span
        >
    </form>
</main>
