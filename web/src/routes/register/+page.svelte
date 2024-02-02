<script lang="ts">
    import { goto } from "$app/navigation";
    import Button from "$lib/components/base/button.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import LogoTextTwoLine from "$lib/components/logo/logo_text_two_line.svelte";
    import { show_toast } from "$lib/stores/toast_store";
    import { users_create, type User, login } from "$lib/stores/user_store";
    import { createForm } from "$lib/vendor/svelte-form-lib";
    import { object, string } from "yup";

    let loading: boolean = false;
    const { form, errors, handleChange, handleSubmit } = createForm<User>({
        initialValues: {
            username: "",
            email: "",
            password: "",
        },
        validationSchema: object<User>({
            username: string().required("Required"),
            email: string().email().required("Required"),
            password: string()
                .min(8, "Must be at least 8 characters long")
                .required("Required"),
        }),
        onSubmit: async (newUser) => {
            loading = true;
            try {
                await users_create(newUser);
            } catch (e) {
                show_toast({
                    icon: "close",
                    type: "error",
                    text: "Error creating user.",
                });
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
                    text: "Error during login.",
                });
            } finally {
                loading = false;
            }
        },
    });
</script>

<main class="flex justify-center">
    <form
        class="login-panel max-w-md shadow-lg rounded-xl p-8 flex flex-col justify-center items-center gap-8 w-[28rem] mt-8"
        on:submit={handleSubmit}
    >
        <LogoTextTwoLine></LogoTextTwoLine>
        <h4 class="text-xl font-semibold">Save your adventures!</h4>
        <div class="space-y-6 w-80">
            <TextField
                name="username"
                label="Username"
                bind:value={$form.username}
                on:change={handleChange}
                error={$errors.username}
            ></TextField>
            <TextField
                name="email"
                label="Email"
                bind:value={$form.email}
                on:change={handleChange}
                error={$errors.email}
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
                {loading}>Register</Button
            >
        </div>
        <span
            >Already have an account? <a
                class="text-blue-500 underline"
                href="/login">Login!</a
            ></span
        >
    </form>
</main>
