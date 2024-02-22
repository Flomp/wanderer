<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import EmptyStateSearch from "$lib/components/empty_states/empty_state_search.svelte";
    import { ms } from "$lib/meilisearch";
    import type { Trail } from "$lib/models/trail";
    import {
        trails,
        trails_search_bounding_box,
    } from "$lib/stores/trail_store";
    import { country_codes } from "$lib/util/country_code_util";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
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

    let L: any;
    let map: Map;
    let gpxLayers: Record<string, GPX> = {};
    let startMarkers: Record<string, Marker> = {};

    let searchDropdownItems: SearchItem[] = [];

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
        const response = await ms.multiSearch({
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
        });

        const trailItems = response.results[0].hits.map((t) => ({
            text: t.name,
            description: `Trail | ${t.location}`,
            value: t,
            icon: "route",
        }));
        const cityItems = response.results[1].hits.map((c) => ({
            text: c.name,
            description: `City | ${
                country_codes[c["country code"] as keyof typeof country_codes]
            }`,
            value: c,
            icon: "city",
        }));

        searchDropdownItems = [...trailItems, ...cityItems];
    }

    function handleSearchClick(item: SearchItem) {
        map.setView([item.value._geo.lat, item.value._geo.lng], 12);
    }

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
            delete startMarkers[deletedTrail.id!];
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
                        startMarkers[trail.id!] = marker;
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
</script>

<main class="grid grid-cols-[400px_1fr]">
    <div
        class="overflow-y-auto overflow-x-hidden flex flex-col items-center gap-4 px-8"
    >
        <Search
            extraClasses="w-full"
            on:update={(e) => search(e.detail)}
            on:click={(e) => handleSearchClick(e.detail)}
            placeholder="Search for trails, places..."
            items={searchDropdownItems}
        ></Search>
        {#if $trails.length == 0}
            <div>
                <div class="w-56 my-4">
                    <EmptyStateSearch></EmptyStateSearch>
                </div>
                <h3 class="text-xl font-semibold text-center">
                    No results found
                </h3>
            </div>
        {/if}
        {#each $trails as trail}
            <a href="/trail/view/{trail.id}">
                <TrailCard
                    {trail}
                    mode="edit"
                    on:mouseenter={() => handleTrailCardMouseEnter(trail)}
                    on:mouseleave={() => handleTrailCardMouseLeave(trail)}
                ></TrailCard>
            </a>
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
