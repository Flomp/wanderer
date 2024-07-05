<script lang="ts">
    import MapWithElevation from "$lib/components/trail/map_with_elevation.svelte";
    import TrailInfoPanel from "$lib/components/trail/trail_info_panel.svelte";
    import { trail } from "$lib/stores/trail_store";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import type { Marker } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import "photoswipe/style.css";
    import { _ } from "svelte-i18n";

    let markers: Marker[];
</script>

<svelte:head>
    <title>{$trail.name} | {$_("map")} | wanderer</title>
</svelte:head>
<main class="grid grid-cols-1 md:grid-cols-[458px_1fr] gap-x-1 gap-y-4">
    <TrailInfoPanel trail={$trail} {markers}></TrailInfoPanel>
    <div id="trail-details" class=" sticky top-0 min-h-[600px]">
        <MapWithElevation trail={$trail} bind:markers></MapWithElevation>
    </div>
</main>

<style>
    @media only screen and (min-width: 768px) {
        #trail-details {
            max-height: calc(100vh - 124px);
        }
    }
</style>
