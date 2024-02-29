<script lang="ts">
    import { goto } from "$app/navigation";
    import SummitLogCard from "$lib/components/summit_log/summit_log_card.svelte";
    import Tabs from "$lib/components/tabs.svelte";
    import TrailDropdown from "$lib/components/trail/trail_dropdown.svelte";
    import WaypointCard from "$lib/components/waypoint/waypoint_card.svelte";
    import { trail } from "$lib/stores/trail_store";
    import { currentUser } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import type { Icon, Map, Marker } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import type { DataSource } from "photoswipe";
    import PhotoSwipeLightbox from "photoswipe/lightbox";
    import "photoswipe/style.css";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";

    const tabs = [
        $_("description"),
        $_("waypoints"),
        $_("photos"),
        $_("summit-book"),
    ];

    let map: Map;

    let activeTab = 0;

    let markers: Marker[] = [];

    let lightbox: PhotoSwipeLightbox;
    let lightboxDataSource: DataSource;

    onMount(async () => {
        const L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");

        map = L.map("map").setView([0, 0], 2);

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);

        const gpxLayer = new L.GPX($trail.expand.gpx_data!, {
            async: true,
            polyline_options: {
                className: "lightblue-theme elevation-polyline",
                opacity: 0.75,
                weight: 5,
            },
            gpx_options: {
                parseElements: ["track"] as any,
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
            .on("loaded", function (e) {
                map.fitBounds(e.target.getBounds());
            })
            .addTo(map);

        for (const waypoint of $trail.expand.waypoints) {
            const marker = createMarkerFromWaypoint(L, waypoint);
            marker.addTo(map);
            markers.push(marker);
        }

        lightboxDataSource = $trail.photos.map((p) => ({
            src: p,
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

    async function toggleMapFullScreen() {
        goto(`/map/trail/${$trail.id!}`);
    }
</script>

<svelte:head>
    <title>{$trail.name} | {$_("trail", { values: { n: 1 } })} | wanderer</title
    >
</svelte:head>

<div
    class="trail-panel max-w-5xl mx-auto border border-input-border rounded-3xl overflow-hidden"
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
            {#if $currentUser && $currentUser.id == $trail.author}
                <TrailDropdown trail={$trail} mode="overview"></TrailDropdown>
            {/if}
        </div>
    </section>
    <section
        class="grid grid-cols-2 sm:grid-cols-4 gap-y-4 justify-around py-8 border-b border-input-border"
    >
        <div class="flex flex-col items-center">
            <span>{$_("distance")}</span>
            <span class="font-semibold text-lg"
                >{formatDistance($trail.distance)}</span
            >
        </div>
        <div class="flex flex-col items-center">
            <span>{$_("elevation-gain")}</span>
            <span class="font-semibold text-lg"
                >{formatElevation($trail.elevation_gain)}</span
            >
        </div>
        <div class="flex flex-col items-center">
            <span>{$_("est-duration")}</span>
            <span class="font-semibold text-lg"
                >{formatTimeHHMM($trail.duration)}</span
            >
        </div>
        {#if $trail.expand.category}
            <div class="flex flex-col items-center">
                <span>{$_("category")}</span>
                <span class="font-semibold text-lg"
                    >{$trail.expand.category.name}</span
                >
            </div>
        {/if}
    </section>
    {#if $trail.tags && $trail.tags.length > 0}
        <hr class="border-separator" />
        <section class="flex p-8 gap-4 text-gray-600">
            {#each $trail.tags as tag}
                <span class="py-2 px-4 border rounded-full">{tag}</span>
            {/each}
        </section>
        <hr class="border-separator" />
    {/if}
    <section class="p-8">
        <Tabs {tabs} bind:activeTab></Tabs>
        <div class="grid grid-cols-1 md:grid-cols-[1fr_18rem] mt-6 gap-8">
            <div>
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
                    <div
                        id="photo-gallery"
                        class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4"
                    >
                        {#each $trail.photos ?? [] as photo, i}
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                            <img
                                class="rounded-xl cursor-pointer hover:scale-105 transition-transform"
                                on:click={() => openGallery(i)}
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
            </div>
            <div class="relative">
                <div class="rounded-xl h-72" id="map">
                    <div class="leaflet-top leaflet-right">
                        <button
                            class="leaflet-control fa fa-maximize rounded-full text-lg bg-white text-black px-[14px] py-2 hover:bg-gray-100"
                            style="cursor: pointer !important"
                            on:click={() => toggleMapFullScreen()}
                        ></button>
                    </div>
                </div>
            </div>
        </div>
    </section>
</div>

<style>
    .trail-panel img {
        object-fit: cover;
    }
</style>
