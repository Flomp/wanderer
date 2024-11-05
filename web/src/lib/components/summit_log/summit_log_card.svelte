<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import "leaflet/dist/leaflet.css";
    import { _ } from "svelte-i18n";
    import Dropdown, { type DropdownItem } from "../base/dropdown.svelte";

    import { browser } from "$app/environment";
    import GPX from "$lib/models/gpx/gpx";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { gpx } from "$lib/vendor/toGeoJSON/toGeoJSON";
    import type { Map } from "leaflet";
    import { onMount } from "svelte";

    let map: Map;
    let L: any;
    let layer: any;

    export let index: number = 0;
    export let log: SummitLog;
    export let mode: "show" | "edit" = "show";

    const dropdownItems: DropdownItem[] = [
        { text: $_("edit"), value: "edit" },
        { text: $_("delete"), value: "delete" },
    ];

    onMount(async () => {
        if (!map) {
            await initMap();
        }
        if (log.expand.gpx_data) {
            showTrailOnMap();
        }
    });

    async function initMap() {
        L = (await import("leaflet")).default;

        map = L.map("mini-map-" + index, {
            zoomControl: false,
            scrollWheelZoom: false,
            dragging: false,
        });
        map.attributionControl.setPrefix(false);
    }

    $: if (log.expand.gpx_data) {
        showTrailOnMap();
    } else {
        removeTrailFromMap();
    }

    async function showTrailOnMap() {
        if (!log.expand.gpx_data || !browser || !map) {
            return;
        }

        const geoJson = gpx(
            new DOMParser().parseFromString(log.expand.gpx_data, "text/xml"),
        );
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

<div class="p-4 my-2 border border-input-border rounded-xl">
    <div class="flex items-center gap-x-4">
        <div
            class="h-24 aspect-square shrink-0 rounded-xl !bg-background"
            class:hidden={!log.expand.gpx_data}
            id="mini-map-{index}"
        ></div>
        <div class="basis-full">
            <div
                class="flex justify-between items-center"
                class:mb-2={log.text}
            >
                <h5 class="font-medium mr-2">
                    {new Date(log.date).toLocaleDateString(undefined, {
                        month: "2-digit",
                        day: "2-digit",
                        year: "numeric",
                        timeZone: "UTC",
                    })}
                </h5>

                {#if mode == "edit"}
                    <Dropdown items={dropdownItems} on:change></Dropdown>
                {/if}
            </div>
            {#if log.distance || log.elevation_gain || log.elevation_loss || log.duration}
                <div
                    class="flex mt-1 gap-x-4 text-sm text-gray-500 flex-wrap mb-2"
                >
                    <span
                        ><i class="fa fa-left-right mr-2"></i>{formatDistance(
                            log.distance,
                        )}</span
                    >
                    <span
                        ><i class="fa fa-clock mr-2"></i>{formatTimeHHMM(
                            log.duration ? log.duration / 60 : undefined,
                        )}</span
                    >
                    <span
                        ><i class="fa fa-arrow-trend-up mr-2"
                        ></i>{formatElevation(log.elevation_gain)}</span
                    >
                    <span
                        ><i class="fa fa-arrow-trend-down mr-2"
                        ></i>{formatElevation(log.elevation_loss)}</span
                    >
                </div>
            {/if}
            <span class="whitespace-pre-wrap">{log.text}</span>
        </div>
    </div>
</div>
