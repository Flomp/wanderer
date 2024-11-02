<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import "leaflet/dist/leaflet.css";
    import { onMount, tick } from "svelte";
    import Modal from "../base/modal.svelte";
    import SummitLogTableRow from "./summit_log_table_row.svelte";

    import type { Map } from "leaflet";
    import { gpx } from "$lib/vendor/toGeoJSON/toGeoJSON";
    import { endIcon, startIcon } from "$lib/util/leaflet_util";
    export let summitLogs: SummitLog[];

    let openModal: () => void;
    let closeModal: () => void;

    let map: Map;
    let L: any;
    let layerGroup: any;

    onMount(async () => {
        L = (await import("leaflet")).default;

        map = L.map("summit-log-table-map");
        map.attributionControl.setPrefix(false);

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);

        layerGroup = L.layerGroup();

        layerGroup.addTo(map);
    });

    async function openMap(log: SummitLog) {
        if (!log.expand.gpx_data) {
            return;
        }
        
        layerGroup.clearLayers();
        const geoJson = gpx(
            new DOMParser().parseFromString(log.expand.gpx_data, "text/xml"),
        );
        const layer = L.geoJson(geoJson, {
            onEachFeature: (feature: any, layer: any) => {
                let startCoords, endCoords;

                if (geoJson.features && geoJson.features.length > 0) {
                    const geometry = geoJson.features[0].geometry;
                    if (geometry.type === "LineString") {
                        startCoords = geometry.coordinates[0];
                        endCoords =
                            geometry.coordinates[
                                geometry.coordinates.length - 1
                            ];
                    } else if (geometry.type === "MultiLineString") {
                        startCoords = (geometry as any).coordinates[0][0];
                        endCoords = (geometry as any).coordinates[
                            geometry.coordinates.length - 1
                        ][
                            (geometry as any).coordinates[
                                geometry.coordinates.length - 1
                            ].length - 1
                        ];
                    }
                }

                if (startCoords && endCoords) {
                    const startMarker = L.marker(
                        [startCoords[1], startCoords[0]],
                        {
                            icon: startIcon(),
                        },
                    );

                    const endMarker = L.marker([endCoords[1], endCoords[0]], {
                        icon: endIcon(),
                    });

                    layerGroup.addLayer(startMarker);
                    layerGroup.addLayer(endMarker)
                }
            },
            filter: (feature: any, layer: any) => {
                return feature.geometry.type !== "Point";
            },
        }).addTo(map);

        layerGroup.addLayer(layer);

        openModal();
        await tick();
        map.fitBounds(layer.getBounds());
        map.invalidateSize();
    }
</script>

<table class="w-full">
    <thead>
        <tr class="text-sm">
            <th class="w-24"></th>
            <th>Date</th>
            <th>Distance</th>
            <th class="whitespace-nowrap">Elevation gain</th>
            <th>Duration</th>
            <th></th>
        </tr>
    </thead>
    <tbody>
        {#each summitLogs as log, i}
            <SummitLogTableRow index={i} {log} on:open={() => openMap(log)}
            ></SummitLogTableRow>
        {/each}
    </tbody>
</table>

<Modal id="summit-log-table-modal" title="" bind:openModal bind:closeModal>
    <div slot="content" id="summit-log-table-map" class="h-96"></div>
</Modal>

<style>
    th {
        padding: 0rem 0.75rem;
    }
</style>
