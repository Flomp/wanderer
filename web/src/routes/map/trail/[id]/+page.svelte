<script lang="ts">
    import SummitLogCard from "$lib/components/summit_log/summit_log_card.svelte";
    import Tabs from "$lib/components/tabs.svelte";
    import WaypointCard from "$lib/components/waypoint/waypoint_card.svelte";
    import { trail } from "$lib/stores/trail_store";
    import { getFileURL } from "$lib/util/file_util";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import type { Map, Marker } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import { onMount } from "svelte";

    let L: any;
    let map: Map;
    let markers: Marker[] = [];

    const tabs = ["Description", "Waypoints", "Photos", "Summit book"];
    let activeTab = 0;

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");
        //@ts-ignore
        await import("$lib/vendor/leaflet-elevation/src/index.js");

        map = L.map("map").setView([$trail.lat, $trail.lon], 14);

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);

        const elevation_options = {
            height: 200,

            theme: "lightblue-theme",
            detached: true,
            elevationDiv: "#elevation",
            closeBtn: false,
            followMarker: true,
            autofitBounds: true,
            imperial: false,
            reverseCoords: false,
            acceleration: false,
            slope: true,
            speed: false,
            altitude: true,
            time: true,
            distance: true,
            // Summary track info style: "inline" || "multiline" || false
            summary: false,
            downloadLink: false,
            ruler: false,
            legend: true,
            // Toggle "leaflet-almostover" integration
            almostOver: true,
            // Toggle "leaflet-distance-markers" integration
            distanceMarkers: false,
            // Toggle "leaflet-edgescale" integration
            edgeScale: false,
            // Toggle "leaflet-hotline" integration
            hotline: true,
            // Display track datetimes: true || false
            timestamps: false,
            waypoints: false,
            wptIcons: false,
            wptLabels: false,
            preferCanvas: true,
        };

        const controlElevation = L.control
            .elevation(elevation_options)
            .addTo(map);
        controlElevation.load(getFileURL($trail, $trail.gpx));

        // addGPXLayer($trail);

        for (const waypoint of $trail.expand.waypoints) {
            const marker = createMarkerFromWaypoint(L, waypoint);
            marker.addTo(map);
            markers.push(marker);
        }
    });

    function openMarkerPopup(i: number) {
        markers[i].openPopup();
    }
</script>

<main class="grid grid-cols-1 md:grid-cols-[452px_1fr] gap-x-1">
    <div
        id="trail-details"
        class="md:overflow-y-auto md:overflow-x-hidden flex flex-col items-stretch gap-4 rounded-xl"
    >
        <section class="relative h-80">
            <img
                class="w-full h-80 object-cover"
                src={getFileURL($trail, $trail.thumbnail)}
                alt=""
            />
            <div
                class="absolute bottom-0 w-full h-1/2 bg-gradient-to-b from-transparent to-black opacity-50"
            ></div>
            <div
                class="flex absolute flex-wrap justify-between items-end w-full bottom-8 left-0 px-8 gap-y-4"
            >
                <div class="text-white">
                    <h4 class="text-4xl font-bold">
                        {$trail.name}
                    </h4>
                    <h3 class="text-xl mt-4">
                        <i class="fa fa-location-dot mr-3"></i>
                        {$trail.location}
                    </h3>
                </div>
            </div>
        </section>
        <section class="grid grid-cols-2 sm:grid-cols-4 gap-y-4 justify-around">
            <div class="flex flex-col items-center text-sm">
                <span class="text-gray-500">Distance</span>
                <span class="font-semibold"
                    >{formatMeters($trail.distance)}</span
                >
            </div>
            <div class="flex flex-col items-center text-sm">
                <span class="text-gray-500">Elevation gain</span>
                <span class="font-semibold"
                    >{formatMeters($trail.elevation_gain)}</span
                >
            </div>
            <div class="flex flex-col items-center text-sm">
                <span class="text-gray-500">Est. duration</span>
                <span class="font-semibold"
                    >{formatTimeHHMM($trail.duration)}</span
                >
            </div>
            {#if $trail.expand.category}
                <div class="flex flex-col items-center text-sm">
                    <span class="text-gray-500">Category</span>
                    <span class="font-semibold"
                        >{$trail.expand.category.name}</span
                    >
                </div>
            {/if}
        </section>
        <hr class=" border-input-border" />
        <section class="mx-4">
            <Tabs extraClasses="text-sm mb-4" {tabs} bind:activeTab></Tabs>
            {#if activeTab == 0}
                <article class="text-justify whitespace-pre-line text-sm">
                    {$trail.description}
                </article>
            {/if}
            {#if activeTab == 1}
                <ul>
                    {#each $trail.expand.waypoints ?? [] as waypoint, i}
                        <li on:mouseenter={() => openMarkerPopup(i)}>
                            <WaypointCard {waypoint}></WaypointCard>
                        </li>
                    {/each}
                </ul>
            {/if}
            {#if activeTab == 2}
                <div id="photo-gallery" class="">
                    {#each $trail.photos ?? [] as photo, i}
                        <img
                            class="rounded-xl cursor-pointer hover:scale-105 transition-transform"
                            src={photo}
                            alt=""
                        />
                    {/each}
                </div>
            {/if}
            {#if activeTab == 3}
                <ul>
                    {#each $trail.expand.summit_logs ?? [] as log}
                        <li><SummitLogCard {log}></SummitLogCard></li>
                    {/each}
                </ul>
            {/if}
        </section>
    </div>
    <div id="map-container" class="flex flex-col">
        <div id="map" class="rounded-xl z-0 basis-full"></div>
        <div id="elevation"></div>
    </div>
</main>

<style>
    #map-container {
        height: calc(100vh - 180px);
    }
    @media only screen and (min-width: 768px) {
        #map-container,
        #trail-details {
            height: calc(100vh - 124px);
        }
    }
</style>
