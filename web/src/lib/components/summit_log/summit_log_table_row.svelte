<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import "leaflet/dist/leaflet.css";

    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { gpx } from "$lib/vendor/toGeoJSON/toGeoJSON";
    import type { Map } from "leaflet";
    import { createEventDispatcher, onMount } from "svelte";
    import { _ } from "svelte-i18n";

    export let index: number;
    export let log: SummitLog;
    export let showCategory: boolean = false;

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

<tr class="text-center">
    <td>
        <button
            type="button"
            on:click={openMap}
            class="h-20 aspect-square shrink-0 rounded-xl !bg-background hover:!bg-secondary-hover transition-colors"
            class:hidden={!log.expand.gpx_data}
            id="mini-map-{index}"
        ></button>
    </td>
    <td class:py-4={!log.expand.gpx_data}
        >{new Date(log.date).toLocaleDateString(undefined, {
            month: "2-digit",
            day: "2-digit",
            year: "numeric",
            timeZone: "UTC",
        })}</td
    >
    <td>
        {formatDistance(log.distance)}
    </td>

    <td>
        {formatElevation(log.elevation_gain)}
    </td>
    <td>
        {formatElevation(log.elevation_loss)}
    </td>
    <td>
        {formatTimeHHMM(log.duration)}
    </td>
    {#if showCategory}
        <td>
            {$_(
                log.expand.trails_via_summit_logs?.at(0)?.expand.category
                    ?.name ?? "-",
            )}
        </td>
    {/if}
    <td>
        {#if log.text}
            <button on:click={() => (showText = !showText)} class="btn-icon"
                ><i class="fa{showText ? '' : '-regular'} fa-message text-gray-500"
                ></i></button
            >
        {/if}
    </td>
</tr>
{#if showText}
    <tr
        ><td
            class="text-left text-sm whitespace-pre-wrap pb-4"
            colspan={showCategory ? 8 : 7}>{log.text}</td
        ></tr
    >
{/if}
<tr>
    <td colspan={showCategory ? 8 : 7}> <hr class="border-input-border" /> </td>
</tr>
