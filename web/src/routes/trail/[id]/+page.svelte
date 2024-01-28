<script lang="ts">
    import SummitLogCard from "$lib/components/summit_log_card.svelte";
    import Tabs from "$lib/components/tabs.svelte";
    import WaypointCard from "$lib/components/waypoint/waypoint_card.svelte";
    import { trail } from "$lib/stores/trail_store";
    import { getFileURL } from "$lib/util/file_util";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import type { Icon, Marker } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import type { DataSource } from "photoswipe";
    import PhotoSwipeLightbox from "photoswipe/lightbox";
    import "photoswipe/style.css";
    import { onMount } from "svelte";

    const tabs = ["Description", "Waypoints", "Photos", "Summit book"];

    let activeTab = 0;

    let markers: Marker[] = [];

    let lightbox: PhotoSwipeLightbox;
    let lightboxDataSource: DataSource;

    onMount(async () => {
        const L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");

        const map = L.map("map").setView([0, 0], 2);

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);

        const gpxLayer = new L.GPX($trail.gpx!, {
            async: true,
            gpx_options: {
                parseElements: ["track"] as any,
            },
            marker_options: {
                wptIcons: {
                    "": L.AwesomeMarkers.icon({
                        icon: "circle",
                        prefix: "fa",
                        markerColor: "cadetblue",
                        iconColor: "white",
                    }) as Icon,
                    Summit: L.AwesomeMarkers.icon({
                        icon: "mountain",
                        prefix: "fa",
                        markerColor: "cadetblue",
                        iconColor: "white",
                    }) as Icon,
                },
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
            .on("loaded", function (e) {
                map.fitBounds(e.target.getBounds());
            })
            .addTo(map);

        for (const waypoint of $trail.expand.waypoints) {
            const marker = createMarkerFromWaypoint(L, waypoint);
            marker.addTo(map);
            markers.push(marker);
        }

        lightboxDataSource = $trail.photos.map((v) => ({
            src: getFileURL($trail, v),
        }));
        lightbox = new PhotoSwipeLightbox({
            dataSource: lightboxDataSource,
            pswpModule: async () => await import("photoswipe"),
        });
        lightbox.init();

        lightbox.on("beforeOpen", () => {
            const pswp = lightbox.pswp;
            const ds = pswp?.options?.dataSource;
            if (Array.isArray(ds)) {
                for (let idx = 0, len = ds.length; idx < len; idx++) {
                    const item = ds[idx];
                    const img = new Image();
                    img.onload = () => {
                        item.width = img.naturalWidth;
                        item.height = img.naturalHeight;
                        pswp?.refreshSlideContent(idx);
                    };
                    img.src = item.src as string;
                }
            }
        });
    });

    function openMarkerPopup(i: number) {
        markers[i].openPopup();
    }

    function openGallery(idx: number) {
        lightbox?.loadAndOpen(idx, lightboxDataSource);
    }
</script>

<div
    class="trail-panel max-w-5xl mx-auto shadow-2xl rounded-3xl overflow-hidden"
>
    <section class="relative h-80">
        <img
            class="w-full h-80"
            src={getFileURL($trail, $trail.thumbnail)}
            alt=""
        />
        <div
            class="absolute bottom-0 w-full h-1/2 bg-gradient-to-b from-transparent to-black opacity-50"
        ></div>
        <div class="absolute text-white bottom-8 left-8">
            <h4 class="text-4xl font-bold">
                {$trail.name}
            </h4>
            <h3 class="text-xl mt-4">
                <i class="fa fa-location-dot mr-3"></i>
                {$trail.location}
            </h3>
        </div>
    </section>
    <section
        class="grid grid-cols-2 sm:grid-cols-4 gap-y-4 justify-around py-8"
    >
        <div class="flex flex-col items-center">
            <span>Distance</span>
            <span class="font-semibold text-lg"
                >{formatMeters($trail.distance)}</span
            >
        </div>
        <div class="flex flex-col items-center">
            <span>Elevation gain</span>
            <span class="font-semibold text-lg"
                >{formatMeters($trail.elevation_gain)}</span
            >
        </div>
        <div class="flex flex-col items-center">
            <span>Est. duration</span>
            <span class="font-semibold text-lg"
                >{formatTimeHHMM($trail.duration)}</span
            >
        </div>
        <div class="flex flex-col items-center">
            <span>Category</span>
            <span class="font-semibold text-lg"
                >{$trail.expand.category.name}</span
            >
        </div>
    </section>
    {#if $trail.tags && $trail.tags.length > 0}
        <hr />
        <section class="flex p-8 gap-4 text-gray-600">
            {#each $trail.tags as tag}
                <span class="py-2 px-4 border rounded-full">{tag}</span>
            {/each}
        </section>
        <hr />
    {/if}
    <section class="p-8">
        <Tabs {tabs} bind:activeTab></Tabs>
        <div class="flex flex-wrap md:flex-nowrap mt-6 gap-8">
            <div class="basis-full">
                {#if activeTab == 0}
                    <article class="text-justify whitespace-pre-line">
                        {$trail.description}
                    </article>
                {/if}
                {#if activeTab == 1}
                    <ul>
                        {#each $trail.expand.waypoints as waypoint, i}
                            <li on:mouseenter={() => openMarkerPopup(i)}>
                                <WaypointCard {waypoint}></WaypointCard>
                            </li>
                        {/each}
                    </ul>
                {/if}
                {#if activeTab == 2}
                    <div
                        id="photo-gallery"
                        class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4"
                    >
                        {#each $trail.photos as photo, i}
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                            <img
                                class="rounded-xl cursor-pointer hover:scale-105 transition-transform"
                                on:click={() => openGallery(i)}
                                src={getFileURL($trail, photo)}
                                alt=""
                            />
                        {/each}
                    </div>
                {/if}
                {#if activeTab == 3}
                    {#each $trail.expand.summit_logs as log}
                        <SummitLogCard {log}></SummitLogCard>
                    {/each}
                {/if}
            </div>
            <div
                class="rounded-xl h-72 basis-full md:basis-72 shrink-0"
                id="map"
            ></div>
        </div>
    </section>
</div>

<style>
    .trail-panel img {
        object-fit: cover;
    }
</style>
