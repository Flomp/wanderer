<script lang="ts">
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import type { Trail } from "$lib/models/trail";
    import {
        trails,
        trails_search_bounding_box,
    } from "$lib/stores/trail_store";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
    import type {
        GPX,
        Icon,
        LatLng,
        LayerGroup,
        LeafletEvent,
        Map,
        Marker,
    } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import { onMount } from "svelte";

    let L: any;
    let map: Map;
    let gpxLayers: Record<string, GPX> = {};

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");

        map = L.map("map").setView([0, 0], 4);

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);

        map.on("moveend", async (e) => {
            const bounds = map.getBounds();
            await searchTrails(bounds.getNorthEast(), bounds.getSouthWest());
        });

        navigator.geolocation.getCurrentPosition(
            (position) => {
                const lat = position.coords.latitude;
                const lon = position.coords.longitude;

                // map.setView([lat, lon], 15);
            },
            (error) => {
                console.error("Error getting user location:", error);
            },
        );
    });

    async function searchTrails(northEast: LatLng, southWest: LatLng) {
        const changes = await trails_search_bounding_box(northEast, southWest);

        for (const newTrail of changes.added) {
            const gpxLayer = await addGPXLayer(newTrail);

            if (gpxLayer) {
                gpxLayers[newTrail.id!] = gpxLayer;
            }
        }

        for (const deletedTrail of changes.deleted) {
            map.removeLayer(gpxLayers[deletedTrail.id!]);
            delete gpxLayers[deletedTrail.id!];
        }
    }

    function addGPXLayer(trail: Trail) {
        return new Promise<GPX>(function (resolve, reject) {
            if (!trail.expand.gpx_data) {
                reject();
            }
            const gpxLayer = new L.GPX(trail.expand.gpx_data!, {
                async: true,
                gpx_options: {
                    parseElements: ["track"],
                },
                marker_options: {
                    startIcon: L.AwesomeMarkers.icon({
                        icon: "circle-half-stroke",
                        prefix: "fa",
                        markerColor: "cadetblue",
                        iconColor: "white",
                    }) as Icon,
                    startIconUrl: "",
                    endIconUrl: "",
                    shadowUrl: "",
                },
            })
                .on("addpoint", function (e: any) {
                    if (e.point_type === "start") {
                        const marker: Marker = e.point as Marker;
                        marker.bindPopup(
                            `<a href="/trail/view/${trail.id}">
    <li class="flex items-center gap-4 cursor-pointer text-black">
        <div class="shrink-0"><img class="h-14 w-14 object-cover rounded-xl" src="${
            trail.thumbnail
        }" alt="">
        </div>
        <div>
            <h4 class="font-semibold text-lg">${trail.name}</h4>
            <h5><i class="fa fa-location-dot mr-3"></i>${trail.location}</h5>
            <div class="flex mt-2 gap-4 text-sm text-gray-500"><span class="shrink-0"><i
                        class="fa fa-left-right mr-2"></i>${formatMeters(
                            trail.distance,
                        )}</span> <span class="shrink-0"><i class="fa fa-up-down mr-2"></i>${formatMeters(
                            trail.elevation_gain,
                        )}</span> <span class="shrink-0"><i class="fa fa-clock mr-2"></i>${formatTimeHHMM(
                            trail.duration,
                        )}</span></div>
        </div>
    </li>
</a>`,
                        );

                        // marker.on("mouseover", (e) => {
                        //     marker.openPopup();
                        // });
                        // marker.on("mouseout", (e) => {
                        //     marker.closePopup();
                        // });
                    }
                })
                .on("loaded", function (e: LeafletEvent) {
                    resolve(gpxLayer);
                })
                .on("error", reject)
                .addTo(map);
        });
    }
</script>

<main class="grid grid-cols-[400px_1fr]">
    <div
        class="overflow-y-scroll overflow-x-hidden flex flex-col items-center gap-4 px-8"
    >
        {#each $trails as trail}
            <TrailCard {trail} mode="edit"></TrailCard>
        {/each}
    </div>
    <div class="rounded-xl" id="map"></div>
</main>

<style>
    #map {
        height: calc(100vh - 124px);
    }

    :global(.leaflet-popup-content) {
        width: max-content !important;
        max-width: 100%;
    }
</style>
