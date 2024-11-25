<script lang="ts">
    import MapWithElevationMaplibre from "$lib/components/trail/map_with_elevation_maplibre.svelte";
    import TrailInfoPanel from "$lib/components/trail/trail_info_panel.svelte";
    import { trail } from "$lib/stores/trail_store";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import * as M from "maplibre-gl";
    import "photoswipe/style.css";
    import { _ } from "svelte-i18n";

    let markers: M.Marker[];
</script>

<svelte:head>
    <title>{$trail.name} | {$_("map")} | wanderer</title>
</svelte:head>
<main class="grid grid-cols-1 md:grid-cols-[458px_1fr] gap-x-1 gap-y-4">
    <TrailInfoPanel trail={$trail} {markers}></TrailInfoPanel>
    <div id="trail-details" class="sticky top-[62px]">
        <MapWithElevationMaplibre trail={$trail} bind:markers></MapWithElevationMaplibre>
    </div>
</main>

<style>
    @media only screen and (min-width: 768px) {
        #trail-details {
            height: calc(100vh - 124px);
        }
    }
</style>
