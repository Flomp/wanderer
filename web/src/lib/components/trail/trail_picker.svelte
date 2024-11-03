<script lang="ts">
    import { gpx } from "$lib/vendor/toGeoJSON/toGeoJSON";
    import "leaflet/dist/leaflet.css";

    import type { Map } from "leaflet";
    import { createEventDispatcher, onMount, tick } from "svelte";
    export let trailFile: File | null;
    export let trailData: string | undefined;
    export let label: string = "";

    let map: Map;
    let L: any;
    let layer: any;

    const dispatch = createEventDispatcher();

    $: if (trailData !== undefined) {
        showTrailOnMap();
    } else {
        removeTrailFromMap();
    }

    onMount(async () => {
        if (!map) {
            await initMap();
        }
    });

    async function initMap() {
        L = (await import("leaflet")).default;

        map = L.map("trail-picker-map", {
            zoomControl: false,
            scrollWheelZoom: false,
            dragging: false,
        });
        map.attributionControl.setPrefix(false);
    }

    function openTrailBrowser() {
        if (trailData) {
            trailFile = null;
            trailData = undefined;

            dispatch("change", null)
        } else {
            document.getElementById("trail-input")!.click();
        }
    }

    async function handleTrailSelection(files?: FileList | null) {
        if (!files) {
            files = (document.getElementById("trail-input") as HTMLInputElement)
                .files;
        }

        if (!files) {
            return;
        }

        trailFile = files.item(0);
        trailData = await trailFile?.text();

        dispatch("change", trailData);
    }

    async function showTrailOnMap() {
        if (!trailData) {
            return;
        }

        const geoJson = gpx(
            new DOMParser().parseFromString(trailData, "text/xml"),
        );
        if (layer) {
            map.removeLayer(layer);
        }
        layer = L.geoJson(geoJson, {
            filter: (feature: any, layer: any) => {
                return feature.geometry.type !== "Point";
            },
        }).addTo(map);
        map.fitBounds(layer.getBounds());
        map.invalidateSize();
    }

    function removeTrailFromMap() {
        if (layer) {
            map?.removeLayer(layer);
        }
    }
</script>

<div>
    {#if label.length}
        <p class="text-sm font-medium pb-1">
            {label}
        </p>
    {/if}
    <div class="flex gap-x-4" role="dialog">
        <input
            type="file"
            id="trail-input"
            accept=".gpx,.GPX,.tcx,.TCX,.kml,.KML,.fit,.FIT"
            multiple={false}
            style="display: none;"
            on:change={() => handleTrailSelection()}
        />
        <button
            type="button"
            on:click={openTrailBrowser}
            class="h-28 aspect-square rounded-xl !bg-background border border-input-border-focus hover:!bg-secondary-hover group"
            id="trail-picker-map"
        >
            {#if !trailFile && !trailData}
                <i class="fa fa-plus text-lg"></i>
            {:else}
                <i
                    class="fa fa-trash text-red-500 text-lg hidden group-hover:block relative"
                    style="z-index: 1000"
                ></i>
            {/if}
        </button>
    </div>
</div>
