<script lang="ts">
    import TrailInfoPanel from "$lib/components/trail/trail_info_panel.svelte";
    import { trail } from "$lib/stores/trail_store";
    import { _ } from "svelte-i18n";
    import { page } from "$app/state";

    let { data } = $props();
</script>

{#if data.trail}
    <TrailInfoPanel
        activeTab={parseInt(page.url.searchParams.get("t") ?? "0")}
        initTrail={data.trail}
        mode="overview"
    ></TrailInfoPanel>
{/if}
<svelte:head>
    <title>{$trail.name} | {$_("trail", { values: { n: 1 } })} | wanderer</title
    >
</svelte:head>
