<script lang="ts">
    import { page } from "$app/state";
    import Button from "$lib/components/base/button.svelte";
    import Modal from "$lib/components/base/modal.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import Toggle from "$lib/components/base/toggle.svelte";
    import { StravaSchema } from "$lib/models/api/integration_schema.js";
    import { Integration } from "$lib/models/integration.js";
    import {
        integrations_create,
        integrations_update,
    } from "$lib/stores/integration_store.js";
    import { show_toast } from "$lib/stores/toast_store.js";
    import { validator } from "@felte/validator-zod";
    import { createForm } from "felte";
    import { _ } from "svelte-i18n";

    let { data } = $props();

    let integration = $state(data.integration);

    const scope = "read_all,activity:read_all";
    const redirectUri = page.url.href + "/callback/strava";

    let stravaSettingsModal: Modal;
    let stravaToggleValue: boolean = $derived(
        integration?.strava?.active || false,
    );

    const { form, errors, data: d } = createForm({
        initialValues: {
            clientId: data.integration?.strava?.clientId ?? "",
            clientSecret: data.integration?.strava?.clientSecret ?? "",
            routes: data.integration?.strava?.routes ?? true,
            activities: data.integration?.strava?.activities ?? true,
            active: data.integration?.strava?.active ?? false,
        },
        extend: validator({
            schema: StravaSchema,
        }),
        onSubmit: async (form) => {
            stravaSettingsModal.closeModal();

            try {
                if (integration) {
                    integration.strava = {
                        clientId: form.clientId,
                        clientSecret: form.clientSecret,
                        routes: form.routes,
                        activities: form.activities,
                        active: integration.strava?.active ?? false,
                    };
                    integration = await integrations_update(integration);
                } else {
                    const newIntegration = new Integration("", {
                        routes: form.routes,
                        activities: form.activities,
                        clientId: form.clientId,
                        clientSecret: form.clientSecret,
                        active: false,
                    });
                    integration = await integrations_create(newIntegration);
                }
            } catch (e) {
                show_toast({
                    text: $_("error-setting-up-strava-integration"),
                    icon: "close",
                    type: "error",
                });
            }
        },
    });
    async function onStravaToggle(value: boolean) {
        if (!integration?.strava) {
            return;
        }
        if (value) {
            const authUrl = `https://www.strava.com/oauth/authorize?client_id=${integration.strava.clientId}&response_type=code&redirect_uri=${redirectUri}&scope=${scope}&approval_prompt=auto`;
            window.location.href = authUrl;
        } else {
            const deauthUrl = `https://www.strava.com/oauth/deauthorize`;

            const r = await fetch(deauthUrl, {
                method: "POST",
                headers: {
                    Authorization: `Bearer ${integration.strava.accessToken}`,
                },
            });

            if (!r.ok) {
                show_toast({
                    text: $_("error-disabling-strava-integration"),
                    icon: "close",
                    type: "error",
                });
                return;
            }
            integration.strava = {
                clientId: integration.strava.clientId,
                clientSecret: integration.strava.clientSecret,
                routes: integration.strava.routes,
                activities: integration.strava.activities,
                accessToken: undefined,
                refreshToken: undefined,
                expiresAt: undefined,
                active: false,
            };
            try {
                integration = await integrations_update(integration);
            } catch (e) {
                show_toast({
                    text: $_("error-disabling-strava-integration"),
                    icon: "close",
                    type: "error",
                });
            }
        }
    }
</script>

<svelte:head>
    <title>{$_("settings")} | wanderer</title>
</svelte:head>

<h3 class="text-2xl font-semibold">{$_("integrations")}</h3>
<hr class="mt-4 mb-6 border-input-border" />

<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
    <div class="border border-input-border rounded-lg p-4 space-y-4">
        <img
            src="https://upload.wikimedia.org/wikipedia/commons/c/cb/Strava_Logo.svg"
            alt="strava logo"
        />
        <div>
            <h5 class="text-xl font-semibold">strava</h5>
            <p class="text-sm text-gray-500">
                Syncs your strava routes with wanderer in regular intervals.
            </p>
        </div>
        <div class="flex items-center justify-between">
            <button
                class="btn-secondary"
                onclick={() => stravaSettingsModal.openModal()}
                ><i class="fa fa-cogs mr-2"></i>{$_("settings")}</button
            >
            <Toggle
                value={stravaToggleValue}
                onchange={onStravaToggle}
                disabled={!integration?.strava}
            ></Toggle>
        </div>
    </div>
</div>

<Modal
    id="strava-settings-modal"
    size="max-w-lg"
    title={"strava " + $_("settings")}
    bind:this={stravaSettingsModal}
>
    {#snippet content()}
        <form id="strava-settings-form" class="space-y-2" use:form>
            <TextField
                label="Client-ID"
                placeholder="000000"
                name="clientId"
                error={$errors.clientId}
            ></TextField>
            <TextField
                label="Client Secret"
                placeholder="de8b3789bd7116d..."
                name="clientSecret"
                type="password"
                error={$errors.clientSecret}
            ></TextField>
            <div class="flex gap-x-4">
                <Toggle name="routes" label={$_("route", { values: { n: 2 } })}
                ></Toggle>
                <Toggle
                    name="activities"
                    label={$_("activity", { values: { n: 2 } })}
                ></Toggle>
            </div>
        </form>
    {/snippet}
    {#snippet footer()}
        <div class="flex items-center gap-4">
            <button
                class="btn-secondary"
                onclick={() => stravaSettingsModal.closeModal()}
                >{$_("cancel")}</button
            >
            <button
                class="btn-primary"
                form="strava-settings-form"
                type="submit"
                name="save">{$_("save")}</button
            >
        </div>
    {/snippet}</Modal
>
