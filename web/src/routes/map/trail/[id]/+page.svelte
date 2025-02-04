<script lang="ts">
    import MapWithElevationMaplibre from "$lib/components/trail/map_with_elevation_maplibre.svelte";
    import TrailInfoPanel from "$lib/components/trail/trail_info_panel.svelte";
    import { trail } from "$lib/stores/trail_store";
    import * as M from "maplibre-gl";
    import "photoswipe/style.css";
    import { _ } from "svelte-i18n";

    let markers: M.Marker[] = $state([]);
</script>

<svelte:head>
    <title>{$trail.name} | {$_("map")} | wanderer</title>
</svelte:head>
<main class="grid grid-cols-1 md:grid-cols-[458px_1fr] gap-x-1 gap-y-4">
    <div id="panel" class="hidden md:block">
        <TrailInfoPanel initTrail={$trail} {markers}></TrailInfoPanel>
    </div>
    <div id="trail-details">
        <MapWithElevationMaplibre
            trails={[$trail]}
            waypoints={$trail.expand?.waypoints}
            activeTrail={0}
            bind:markers
            showTerrain={true}
        ></MapWithElevationMaplibre>
    </div>
</main>

<style>
    #trail-details, #panel {
        height: calc(100vh);
    }
    @media only screen and (min-width: 768px) {
        #trail-details, #panel {
            height: calc(100vh - 124px);
        }
    }
</style>
