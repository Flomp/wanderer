<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import "leaflet/dist/leaflet.css";

    import GPX from "$lib/models/gpx/gpx";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { gpx } from "$lib/vendor/toGeoJSON/toGeoJSON";
    import type { Map } from "leaflet";
    import { createEventDispatcher, onMount } from "svelte";

    export let index: number;
    export let log: SummitLog;

    let totals: {
        distance: number;
        elevationGain: number;
        duration: number;
    } | null = null;

    let map: Map;

    let showText: boolean = false;

    const dispatch = createEventDispatcher();

    onMount(async () => {
        if (log.expand.gpx_data) {
            await initMap();
        }
    });

    async function initMap() {
        const L = (await import("leaflet")).default;

        map = L.map("mini-map-" + index, {
            zoomControl: false,
            scrollWheelZoom: false,
            dragging: false,
        });
        map.attributionControl.setPrefix(false);

        if (!log.expand.gpx_data || !map) {
            return;
        }

        const gpxObject = await GPX.parse(log.expand.gpx_data);
        if (gpxObject instanceof Error) {
            throw gpxObject;
        }

        totals = gpxObject.getTotals();

        const geoJson = gpx(
            new DOMParser().parseFromString(log.expand.gpx_data, "text/xml"),
        );
        const layer = (L as any)
            .geoJson(geoJson, {
                filter: (feature: any, layer: any) => {
                    return feature.geometry.type !== "Point";
                },
            })
            .addTo(map);
        map.fitBounds(layer.getBounds());
        map.invalidateSize();
    }

    function openMap() {
        dispatch("open", log);
    }
</script>

<tr class="h-24 text-center">
    <td>
        <button
            type="button"
            on:click={openMap}
            class="h-24 aspect-square shrink-0 rounded-xl !bg-background"
            class:hidden={!log.expand.gpx_data}
            id="mini-map-{index}"
        ></button>
    </td>
    <td
        >{new Date(log.date).toLocaleDateString(undefined, {
            month: "2-digit",
            day: "2-digit",
            year: "numeric",
            timeZone: "UTC",
        })}</td
    >
    <td>
        {formatDistance(totals?.distance)}
    </td>

    <td>
        {formatElevation(totals?.elevationGain)}
    </td><td>
        {formatTimeHHMM(totals?.duration)}
    </td>
    <td>
        {#if log.text}
            <button on:click={() => (showText = !showText)} class="btn-icon"
                ><i class="fa{showText ? '' : '-regular'} fa-message"
                ></i></button
            >
        {/if}
    </td>
</tr>
{#if showText}
    <tr
        ><td class="text-left text-sm whitespace-pre-wrap" colspan="6"
            >{log.text}</td
        ></tr
    >
{/if}
