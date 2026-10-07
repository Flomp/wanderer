<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import type { PluginInstance } from "$lib/models/plugin_instance";
    import type { PluginProvider } from "$lib/models/plugin_provider";
    import type { Trail } from "$lib/models/trail";
    import { plugin_instances_index } from "$lib/stores/plugin_instance_store";
    import { canUsePluginConnection } from "$lib/util/plugin_auth_context";
    import { plugins_index } from "$lib/stores/plugin_store";
    import { theme } from "$lib/stores/theme_store";
    import { show_toast } from "$lib/stores/toast_store.svelte";
    import { APIError } from "$lib/util/api_util";
    import { translatePluginAPIError, pluginOperationNeedsReload, pluginOperationCode } from "$lib/util/plugin_error_i18n";
    import { pluginTitle as localizedPluginTitle } from "$lib/util/plugin_i18n";
    import { onMount } from "svelte";
    import { _, locale } from "svelte-i18n";

    interface Props {
        trail?: Trail;
        share?: string;
    }

    let { trail, share }: Props = $props();

    const uid = $props.id();
    const modalId = `send-modal-${uid}`;

    let modal: Modal;
    let loading = $state(false);
    let sending = $state("");
    let modalOpen = $state(false);
    let instances: PluginInstance[] = $state([]);
    let eligible: PluginProvider[] = $state([]);
    let currentTheme: "dark" | "light" = $state("light");

    onMount(() => {
        currentTheme = document.documentElement.classList.contains("dark") ? "dark" : "light";
        const unsubscribe = theme.subscribe(value => { currentTheme = value; });
        const busyRefresh = window.setInterval(() => {
            if (modalOpen && !document.hidden && instances.some(instance => instance.busy)) void loadEligible();
        }, 3000);
        return () => { unsubscribe(); window.clearInterval(busyRefresh); };
    });

    export async function openModal() {
        modalOpen = true;
        modal.openModal();
        await loadEligible();
    }

    async function loadEligible() {
        loading = true;
        eligible = [];
        try {
            const [plugins, currentInstances] = await Promise.all([
                plugins_index(),
                plugin_instances_index(),
            ]);
            instances = currentInstances;
            eligible = plugins.filter(plugin => plugin.status === "available" &&
                (plugin.capabilities ?? []).includes("prepare_trail_send.v1") &&
                instances.some(instance => instance.plugin_id === plugin.id && canUsePluginConnection(plugin, instance)));

        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    }

    function pluginLogo(plugin: PluginProvider): string | undefined {
        if (currentTheme === "dark" && plugin.iconDark) {
            return plugin.iconDark;
        }
        return plugin.icon || undefined;
    }

    function pluginTitle(plugin: PluginProvider): string {
        return localizedPluginTitle(plugin, $locale);
    }

    async function send(plugin: PluginProvider) {
        if (!trail?.id) {
            return;
        }
        sending = plugin.id;
        try {
            const response = await fetch("/api/v1/plugin-system/trail-send", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ pluginId: plugin.id, trailId: trail.id, share }),
            });
            if (!response.ok) {
                const body = await response.json().catch(() => undefined);
                throw new APIError(response.status, "Plugin operation failed.", body);
            }
            show_toast({
                type: "success",
                icon: "check",
                text: $_("trail-sent"),
            });
            modalOpen = false;
            modal.closeModal();
        } catch (e) {
            if (pluginOperationNeedsReload(e) || pluginOperationCode(e) === "plugin_instance_busy") await loadEligible();
            show_toast({
                type: "error",
                icon: "close",
                text: translatePluginAPIError(e, $_("error-sending-trail")),
            });
        } finally {
            sending = "";
            await plugin_instances_index().catch(() => {});
        }
    }
</script>

<Modal id={modalId} title={$_("send-to")} size="md:min-w-sm" bind:this={modal}>
    {#snippet content()}
        {#if loading}
            <div class="flex justify-center p-4">
                <div class="spinner light:spinner-dark"></div>
            </div>
        {:else if eligible.length === 0}
            <p class="text-sm text-gray-500">{$_("no-send-plugins")}</p>
            <a class="btn-secondary mt-4 inline-block" href="/settings/plugins"
                >{$_("plugins")}</a
            >
        {:else}
            <div class="flex flex-wrap items-center gap-4">
                {#each eligible as plugin (plugin.id)}
                    {@const busy = instances.find(instance => instance.plugin_id === plugin.id)?.busy}
                    <button
                        class="btn-secondary"
                        aria-label={pluginTitle(plugin)}
                        disabled={sending !== "" || busy}
                        onclick={() => send(plugin)}
                    >
                        {#if sending === plugin.id}
                            <div class="spinner light:spinner-dark h-20"></div>
                        {:else if pluginLogo(plugin)}
                            <img
                                class="h-20"
                                src={pluginLogo(plugin)}
                                alt={pluginTitle(plugin)}
                            />
                        {:else}
                            <span>{pluginTitle(plugin)}</span>
                        {/if}
                        {#if busy}<span class="block text-xs mt-2">{$_("plugin-operation-running")}</span>{/if}
                    </button>
                {/each}
            </div>
        {/if}
    {/snippet}
    {#snippet footer()}
        <div class="flex items-center gap-4">
            <button class="btn-secondary" onclick={() => { modalOpen = false; modal.closeModal(); }}
                >{$_("cancel")}</button
            >
        </div>
    {/snippet}
</Modal>
