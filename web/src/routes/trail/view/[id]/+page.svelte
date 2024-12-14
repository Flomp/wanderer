<script>
    import TrailInfoPanel from "$lib/components/trail/trail_info_panel.svelte";
    import { trail } from "$lib/stores/trail_store";
    import { _ } from "svelte-i18n";
    import { page } from "$app/stores";

    export let data;
</script>

<TrailInfoPanel
    activeTab={parseInt($page.url.searchParams.get("t") ?? 0)}
    trail={data.trail}
    mode="overview"
></TrailInfoPanel>

<svelte:head>
    <title>{$trail.name} | {$_("trail", { values: { n: 1 } })} | wanderer</title
    >
</svelte:head>
