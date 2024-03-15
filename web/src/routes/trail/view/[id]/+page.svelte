<script>
    import TrailInfoPanel from "$lib/components/trail/trail_info_panel.svelte";
    import { trail } from "$lib/stores/trail_store";
    import { _ } from "svelte-i18n";
</script>

<TrailInfoPanel trail={$trail} mode="overview"></TrailInfoPanel>

<svelte:head>
    <title>{$trail.name} | {$_("trail", { values: { n: 1 } })} | wanderer</title
    >
</svelte:head>
