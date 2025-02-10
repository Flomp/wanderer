<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import { KomootSchema } from "$lib/models/api/integration_schema";
    import type {
        Integration,
        KomootIntegration,
    } from "$lib/models/integration";
    import { validator } from "@felte/validator-zod";
    import { createForm } from "felte";
    import { _ } from "svelte-i18n";

    interface Props {
        integration?: Integration;
        onsave?: (komootIntegration: KomootIntegration) => void;
    }

    let { integration, onsave }: Props = $props();

    let modal: Modal;

    export function openModal() {
        errors.set({})
        modal.openModal();
    }

    const {
        form,
        errors,
        data: d,
    } = createForm({
        initialValues: {
            email: integration?.komoot?.email ?? "",
            password: integration?.komoot?.password ?? "",
            active: integration?.komoot?.active ?? false,
        },
        extend: validator({
            schema: KomootSchema,
        }),
        onSubmit: async (form) => {
            onsave?.(form);
            modal.closeModal();
        },
    });
</script>

<Modal
    id="komoot-settings-modal"
    size="max-w-lg"
    title={"komoot " + $_("settings")}
    bind:this={modal}
>
    {#snippet content()}
        <form id="komoot-settings-form" class="space-y-2" use:form>
            <TextField
                label={$_("email")}
                placeholder="user@example.com"
                name="email"
                error={$errors.email}
            ></TextField>
            <TextField
                label={$_("password")}
                placeholder={integration?.komoot ? `(${$_("unchanged")})` : ""}
                name="password"
                type="password"
                error={$errors.password}
            ></TextField>
        </form>
    {/snippet}
    {#snippet footer()}
        <div class="flex items-center gap-4">
            <button class="btn-secondary" onclick={() => modal.closeModal()}
                >{$_("cancel")}</button
            >
            <button
                class="btn-primary"
                form="komoot-settings-form"
                type="submit"
                name="save">{$_("save")}</button
            >
        </div>
    {/snippet}</Modal
>
