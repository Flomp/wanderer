<script lang="ts">
    import { plugin_oauth_callback } from "$lib/stores/plugin_instance_store";
    import { translatePluginAPIError, translatePluginError } from "$lib/util/plugin_error_i18n";
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";

    let error = $state("");

    onMount(async () => {
        const code = page.url.searchParams.get("code") ?? "";
        const state = page.url.searchParams.get("state") ?? "";
        const providerError = page.url.searchParams.get("error");
        const instanceId = sessionStorage.getItem(`wanderer_plugin_oauth_${state}`) ?? "";
        sessionStorage.removeItem(`wanderer_plugin_oauth_${state}`);

        if (providerError) {
            error = translatePluginError(providerError);
            return;
        }
        if (!code || !state || !instanceId) {
            error = $_("plugin-error-invalid-request");
            return;
        }

        try {
            await plugin_oauth_callback({ instanceId, code, state });
            await goto("/settings/plugins");
        } catch (e) {
            error = translatePluginAPIError(e, $_("error-completing-oauth"));
        }
    });
</script>

<svelte:head>
    <title>{$_("plugins")} | wanderer</title>
</svelte:head>

<div class="mx-auto max-w-md py-12">
    {#if error}
        <h3 class="text-xl font-semibold">{$_("error")}</h3>
        <p class="mt-3 text-sm text-gray-500">{error}</p>
        <a class="btn-secondary mt-6 inline-block" href="/settings/plugins">{$_("plugins")}</a>
    {:else}
        <h3 class="text-xl font-semibold">{$_("plugins")}</h3>
    {/if}
</div>
