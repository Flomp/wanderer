<script lang="ts">
    import SummitLogCard from "$lib/components/summit_log/summit_log_card.svelte";
    import Tabs from "$lib/components/tabs.svelte";
    import MapWithElevation from "$lib/components/trail/map_with_elevation.svelte";
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
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import type { Marker } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import type { DataSource } from "photoswipe/lightbox";
    import PhotoSwipeLightbox from "photoswipe/lightbox";
    import "photoswipe/style.css";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";

    let markers: Marker[];

    let lightbox: PhotoSwipeLightbox;
    let lightboxDataSource: DataSource;

    const tabs = [
        $_("description"),
        $_("waypoints"),
        $_("photos"),
        $_("summit-book"),
    ];
    let activeTab = 0;

    const thumbnail = $trail.photos.length
        ? getFileURL($trail, $trail.photos[$trail.thumbnail])
        : "/imgs/default_thumbnail.webp";

    onMount(() => {
        lightboxDataSource = $trail.photos.map((p) => ({
            src: getFileURL($trail, p),
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

<svelte:head>
    <title>{$trail.name} | {$_("map")} | wanderer</title>
</svelte:head>
<main class="grid grid-cols-1 md:grid-cols-[458px_1fr] gap-x-1 gap-y-4">
    <div
        id="trail-details"
        class="md:overflow-y-auto md:overflow-x-hidden flex flex-col items-stretch gap-4 rounded-xl"
    >
        <section class="relative h-80">
            <img
                class="w-full h-80 object-cover"
                src={thumbnail}
                alt=""
            />
            <div
                class="absolute bottom-0 w-full h-1/2 bg-gradient-to-b from-transparent to-black opacity-50"
            ></div>
            <div
                class="flex absolute justify-between items-end w-full bottom-8 left-0 px-8 gap-y-4"
            >
                <div class="text-white">
                    <h4 class="text-4xl font-bold">
                        {$trail.name}
                    </h4>
                    <div class="flex flex-wrap gap-x-8 gap-y-2  mt-4 mr-8">
                        <h3 class="text-lg">
                            <i class="fa fa-location-dot mr-3"></i>
                            {$trail.location}
                        </h3>
                        <h3 class="text-lg">
                            <i class="fa fa-person-hiking mr-3"></i>
                            {$_($trail.difficulty ?? "?")}
                        </h3>
                    </div>
                </div>
                {#if $currentUser && $currentUser.id == $trail.author}
                    <TrailDropdown trail={$trail} mode="map"></TrailDropdown>
                {/if}
            </div>
        </section>
        <section class="grid grid-cols-2 sm:grid-cols-4 gap-y-4 justify-around">
            <div class="flex flex-col items-center text-sm">
                <span class="text-gray-500">{$_('distance')}</span>
                <span class="font-semibold"
                    >{formatDistance($trail.distance)}</span
                >
            </div>
            <div class="flex flex-col items-center text-sm">
                <span class="text-gray-500">{$_('elevation-gain')}</span>
                <span class="font-semibold"
                    >{formatElevation($trail.elevation_gain)}</span
                >
            </div>
            <div class="flex flex-col items-center text-sm">
                <span class="text-gray-500">{$_('est-duration')}</span>
                <span class="font-semibold"
                    >{formatTimeHHMM($trail.duration)}</span
                >
            </div>
            {#if $trail.expand.category}
                <div class="flex flex-col items-center text-sm">
                    <span class="text-gray-500">{$_('category')}</span>
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
                <div id="photo-gallery" class="space-y-4 mx-4">
                    {#each $trail.photos ?? [] as photo, i}
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
                <ul>
                    {#each $trail.expand.summit_logs ?? [] as log}
                        <li><SummitLogCard {log}></SummitLogCard></li>
                    {/each}
                </ul>
            {/if}
        </section>
    </div>
    <MapWithElevation trail={$trail} bind:markers></MapWithElevation>
</main>

<style>
    @media only screen and (min-width: 768px) {
        #trail-details {
            height: calc(100vh - 124px);
        }
    }
</style>
