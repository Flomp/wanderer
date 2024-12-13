<script>
    import TrailInfoPanel from "$lib/components/trail/trail_info_panel.svelte";
    import { trail } from "$lib/stores/trail_store";
    import { _ } from "svelte-i18n";

    export let data;
</script>

<TrailInfoPanel trail={data.trail} mode="overview"></TrailInfoPanel>

<svelte:head>
    <title>{$trail.name} | {$_("trail", { values: { n: 1 } })} | wanderer</title
    >
</svelte:head>
