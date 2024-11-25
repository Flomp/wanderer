<script>
    import MapWithElevationMaplibre from "$lib/components/trail/map_with_elevation_maplibre.svelte";
    import { onMount } from "svelte";
    export let data;

</script>

<MapWithElevationMaplibre trail={data.trail}></MapWithElevationMaplibre>
