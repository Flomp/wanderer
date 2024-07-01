<script lang="ts">
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import EmptyStateSearch from "$lib/components/empty_states/empty_state_search.svelte";
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import TrailFilterPanel from "$lib/components/trail/trail_filter_panel.svelte";
    import type { Settings } from "$lib/models/settings";
    import type {
        Trail,
        TrailBoundingBox,
        TrailFilter,
    } from "$lib/models/trail";
    import { categories } from "$lib/stores/category_store";
    import {
        trails,
        trails_search_bounding_box,
    } from "$lib/stores/trail_store";
    import { currentUser } from "$lib/stores/user_store";
    import { country_codes } from "$lib/util/country_code_util";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import type {
        GPX,
        Icon,
        LatLng,
        LatLngBoundsExpression,
        LeafletEvent,
        Map,
        Marker,
    } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import { slide } from "svelte/transition";
    let L: any;
    let map: Map;
    let gpxLayers: Record<string, GPX> = {};
    let startMarkers: Record<string, Marker> = {};

    let searchDropdownItems: SearchItem[] = [];

    let showFilter: boolean = false;
    let showMap: boolean = true;

    const filter: TrailFilter = $page.data.filter;
    const maxBoundingBox: TrailBoundingBox = $page.data.boundingBox;
    const settings: Settings = $page.data.settings;

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");

        map = L.map("map").setView([0, 0], 4);
        map.attributionControl.setPrefix(false);

        const baseLayer = L.tileLayer(
            "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
            {
                attribution: "© OpenStreetMap contributors",
            },
        );

        const topoLayer = L.tileLayer(
            "https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png ",
            {
                attribution: "© OpenStreetMap contributors",
            },
        );

        const baseMaps = {
            OpenStreetMaps: baseLayer,
            OpenTopoMaps: topoLayer,
        };

        L.control.layers(baseMaps).addTo(map);

        switch (localStorage.getItem("layer")) {
            case "OpenTopoMaps":
                topoLayer.addTo(map);
                break;
            default:
                baseLayer.addTo(map);
                break;
        }

        map.on("baselayerchange", function (e) {
            localStorage.setItem("layer", e.name);
        });

        map.on("moveend", async (e) => {
            const bounds = map.getBounds();
            await searchTrails(bounds.getNorthEast(), bounds.getSouthWest());

            $page.url.searchParams.set("tl_lat", bounds.getNorth().toString());
            $page.url.searchParams.set("tl_lon", bounds.getEast().toString());
            $page.url.searchParams.set("br_lat", bounds.getSouth().toString());
            $page.url.searchParams.set("br_lon", bounds.getWest().toString());

            goto(`?${$page.url.searchParams.toString()}`);
        });

        const lat = $page.url.searchParams.get("lat");
        const lon = $page.url.searchParams.get("lon");

        const tl_lat = $page.url.searchParams.get("tl_lat");
        const tl_lon = $page.url.searchParams.get("tl_lon");
        const br_lat = $page.url.searchParams.get("br_lat");
        const br_lon = $page.url.searchParams.get("br_lon");

        if (lat && lon) {
            map.setView([parseFloat(lat), parseFloat(lon)], 12);
        } else if (tl_lat && tl_lon && br_lat && br_lon) {
            const boundingBox: LatLngBoundsExpression = [
                [parseFloat(tl_lat), parseFloat(tl_lon)],
                [parseFloat(br_lat), parseFloat(br_lon)],
            ];

            map.fitBounds(boundingBox);
        } else if (settings && settings.mapFocus == "trails") {
            const boundingBox: LatLngBoundsExpression = [
                [maxBoundingBox.max_lat, maxBoundingBox.min_lon],
                [maxBoundingBox.min_lat, maxBoundingBox.max_lon],
            ];
            map.fitBounds(boundingBox);
        } else if (
            settings &&
            settings.mapFocus == "location" &&
            settings.location
        ) {
            map.setView([settings.location.lat, settings.location.lon], 12);
        } else {
            navigator.geolocation.getCurrentPosition(
                (position) => {
                    const lat = position.coords.latitude;
                    const lon = position.coords.longitude;

                    map.setView([lat, lon], 13);
                },
                (error) => {
                    console.error("Error getting user location:", error);
                },
            );
        }
    });

    async function search(q: string) {
        const r = await fetch("/api/v1/search/multi", {
            method: "POST",
            body: JSON.stringify({
                queries: [
                    {
                        indexUid: "trails",
                        q: q,
                        limit: 3,
                    },
                    {
                        indexUid: "cities500",
                        q: q,
                        limit: 3,
                    },
                ],
            }),
        });

        const response = await r.json();

        const trailItems = response.results[0].hits.map(
            (t: Record<string, any>) => ({
                text: t.name,
                description: `Trail | ${t.location}`,
                value: t,
                icon: "route",
            }),
        );
        const cityItems = response.results[1].hits.map(
            (c: Record<string, any>) => ({
                text: c.name,
                description: `City ${c.division ? `| ${c.division} ` : ""}| ${
                    country_codes[
                        c["country code"] as keyof typeof country_codes
                    ]
                }`,
                value: c,
                icon: "city",
            }),
        );

        searchDropdownItems = [...trailItems, ...cityItems];
    }

    function handleSearchClick(item: SearchItem) {
        map.setView([item.value._geo.lat, item.value._geo.lng], 12);
    }

    async function searchTrails(northEast: LatLng, southWest: LatLng) {
        const changes = await trails_search_bounding_box(
            northEast,
            southWest,
            filter,
        );

        for (const newTrail of changes.added) {
            const gpxLayer = await addGPXLayer(newTrail);

            if (gpxLayer) {
                gpxLayers[newTrail.id!] = gpxLayer;
            }
        }

        for (const deletedTrail of changes.deleted) {
            map.removeLayer(gpxLayers[deletedTrail.id!]);
            delete gpxLayers[deletedTrail.id!];
            delete startMarkers[deletedTrail.id!];
        }
    }

    function addGPXLayer(trail: Trail) {
        return new Promise<GPX>(function (resolve, reject) {
            if (!trail.expand.gpx_data) {
                reject();
            }

            const thumbnail = trail.photos.length
                ? getFileURL(trail, trail.photos[trail.thumbnail])
                : "/imgs/default_thumbnail.webp";
            const gpxLayer = new L.GPX(trail.expand.gpx_data!, {
                async: true,
                polyline_options: {
                    className: "lightblue-theme elevation-polyline",
                    weight: 5,
                },
                gpx_options: {
                    parseElements: ["track", "route"],
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
                        startMarkers[trail.id!] = marker;
                        marker.bindPopup(
                            `<a href="map/trail/${trail.id}">
    <li class="flex items-center gap-4 cursor-pointer text-black">
        <div class="shrink-0"><img class="h-14 w-14 object-cover rounded-xl" src="${thumbnail}" alt="">
        </div>
        <div>
            <h4 class="font-semibold text-lg">${trail.name}</h4>
            <div class="flex gap-x-4">
            ${trail.location ? `<h5><i class="fa fa-location-dot mr-2"></i>${trail.location}</h5>` : ""}
            <h5><i class="fa fa-gauge mr-2"></i>${$_(trail.difficulty as string)}</h5>
            </div>
            <div class="flex mt-2 gap-4 text-sm text-gray-500"><span class="shrink-0"><i
                        class="fa fa-left-right mr-2"></i>${formatDistance(
                            trail.distance,
                        )}</span> <span class="shrink-0"><i class="fa fa-up-down mr-2"></i>${formatElevation(
                            trail.elevation_gain,
                        )}</span> <span class="shrink-0"><i class="fa fa-clock mr-2"></i>${formatTimeHHMM(
                            trail.duration,
                        )}</span></div>
        </div>
    </li>
</a>`,
                        );
                    }
                })
                .on("loaded", function (e: LeafletEvent) {
                    resolve(gpxLayer);
                })
                .on("error", reject)
                .addTo(map);
        });
    }

    function handleTrailCardMouseEnter(trail: Trail) {
        const marker: Marker = startMarkers[trail.id!];
        marker?.openPopup();
    }

    function handleTrailCardMouseLeave(trail: Trail) {
        const marker: Marker = startMarkers[trail.id!];
        marker?.closePopup();
    }

    async function handleFilterUpdate(filter: TrailFilter) {
        const bounds = map.getBounds();
        await searchTrails(bounds.getNorthEast(), bounds.getSouthWest());
    }
</script>

<svelte:head>
    <title>{$_("map")} | wanderer</title>
</svelte:head>
<main class="grid grid-cols-1 md:grid-cols-[400px_1fr]">
    <div
        id="trail-list"
        class="md:overflow-y-auto md:overflow-x-hidden flex flex-col items-stretch gap-4 px-3 md:px-8"
    >
        <div class="sticky top-0 z-10 bg-background pb-4 space-y-4">
            <div class="flex items-center gap-2 md:gap-4">
                <Search
                    extraClasses="w-full"
                    on:update={(e) => search(e.detail)}
                    on:click={(e) => handleSearchClick(e.detail)}
                    placeholder="{$_('search-for-trails-places')}..."
                    items={searchDropdownItems}
                ></Search>
                <button
                    class="btn-icon"
                    on:click={() => (showFilter = !showFilter)}
                    ><i class="fa fa-sliders"></i></button
                >
                <button
                    class="btn-icon md:hidden"
                    on:click={() => (showMap = !showMap)}
                    ><i
                        class="fa-regular fa-{showMap
                            ? 'rectangle-list'
                            : 'map'}"
                    ></i></button
                >
            </div>
            {#if showFilter}
                <div in:slide out:slide>
                    <TrailFilterPanel
                        categories={$categories}
                        showTrailSearch={false}
                        showCitySearch={false}
                        {filter}
                        on:update={(e) => handleFilterUpdate(e.detail)}
                    ></TrailFilterPanel>
                </div>
            {/if}
        </div>

        {#if !showFilter && (!showMap || (browser && window.innerWidth >= 768))}
            {#if $trails.length == 0}
                <EmptyStateSearch></EmptyStateSearch>
            {/if}
            {#each $trails as trail}
                <a href="map/trail/{trail.id}">
                    <TrailCard
                        {trail}
                        on:mouseenter={() => handleTrailCardMouseEnter(trail)}
                        on:mouseleave={() => handleTrailCardMouseLeave(trail)}
                    ></TrailCard>
                </a>
            {/each}
        {/if}
    </div>
    <div
        id="map"
        class="rounded-xl z-0"
        class:hidden={!showMap && browser && window.innerWidth < 768}
    ></div>
</main>

<style>
    #map {
        height: calc(100vh - 180px);
    }
    @media only screen and (min-width: 768px) {
        #map,
        #trail-list {
            height: calc(100vh - 124px);
        }
    }
</style>
