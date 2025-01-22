<script lang="ts">
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { page } from "$app/state";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import SkeletonCard from "$lib/components/base/skeleton_card.svelte";
    import EmptyStateSearch from "$lib/components/empty_states/empty_state_search.svelte";
    import MapWithElevationMaplibre from "$lib/components/trail/map_with_elevation_maplibre.svelte";
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import TrailFilterPanel from "$lib/components/trail/trail_filter_panel.svelte";
    import type { Settings } from "$lib/models/settings";
    import type {
        Trail,
        TrailBoundingBox,
        TrailFilter,
    } from "$lib/models/trail";
    import { categories } from "$lib/stores/category_store";
    import { trails_search_bounding_box } from "$lib/stores/trail_store";
    import { country_codes } from "$lib/util/country_code_util";
    import * as M from "maplibre-gl";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import { slide } from "svelte/transition";

    let trails: Trail[] = $state([]);

    let map: M.Map | undefined = $state();
    let mapWithElevation: MapWithElevationMaplibre | undefined = $state();
    let searchDropdownItems: SearchItem[] = $state([]);

    let showFilter: boolean = $state(false);
    let showMap: boolean = $state(true);

    const filter: TrailFilter = page.data.filter;
    const maxBoundingBox: TrailBoundingBox = page.data.boundingBox;
    const settings: Settings = page.data.settings;

    const MIN_ZOOM = 6;

    let loading: boolean = $state(true);

    onMount(async () => {});

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
        map?.setCenter([item.value._geo.lng, item.value._geo.lat]);
        map?.setZoom(14);
    }

    async function searchTrails(northEast: M.LngLat, southWest: M.LngLat) {
        loading = true;
        const changes = await trails_search_bounding_box(
            northEast,
            southWest,
            filter,
            (map?.getZoom() ?? 0) > MIN_ZOOM,
        );
        trails = changes.trails;
        loading = false;
    }

    function handleTrailCardMouseEnter(trail: Trail) {
        mapWithElevation?.togglePopup(trail.id!, false);
    }

    function handleTrailCardMouseLeave(trail: Trail) {
        mapWithElevation?.togglePopup(trail.id!, true);
    }

    async function handleFilterUpdate(filter: TrailFilter) {
        if(!map) {
            return;
        }
        const bounds = map.getBounds();
        await searchTrails(bounds.getNorthEast(), bounds.getSouthWest());
    }

    async function handleMapMove() {
        if(!map) {
            return;
        }
        const bounds = map.getBounds();

        const normalizedBounds = {
            southWest: new M.LngLat(
                ((((bounds.getSouthWest().lng + 180) % 360) + 360) % 360) - 180,
                bounds.getSouthWest().lat,
            ),
            northEast: new M.LngLat(
                ((((bounds.getNorthEast().lng + 180) % 360) + 360) % 360) - 180,
                bounds.getNorthEast().lat,
            ),
        };        

        await searchTrails(
            normalizedBounds.northEast,
            normalizedBounds.southWest,
        );

        page.url.searchParams.set("tl_lat", bounds.getNorth().toString());
        page.url.searchParams.set("tl_lon", bounds.getEast().toString());
        page.url.searchParams.set("br_lat", bounds.getSouth().toString());
        page.url.searchParams.set("br_lon", bounds.getWest().toString());

        goto(`?${page.url.searchParams.toString()}`);
    }

    function handleMapInit() {
        if (
            page.url.searchParams.has("tl_lat") &&
            page.url.searchParams.has("tl_lon") &&
            page.url.searchParams.has("br_lat") &&
            page.url.searchParams.has("br_lon")
        ) {
            const boundingBox: M.LngLatBoundsLike = [
                [
                    parseFloat(page.url.searchParams.get("br_lon")!),
                    parseFloat(page.url.searchParams.get("tl_lat")!),
                ],
                [
                    parseFloat(page.url.searchParams.get("tl_lon")!),
                    parseFloat(page.url.searchParams.get("br_lat")!),
                ],
            ];
            map?.fitBounds(boundingBox, { animate: false });
        } else if (
            page.url.searchParams.has("lat") &&
            page.url.searchParams.has("lon")
        ) {
            const lat = page.url.searchParams.get("lat");
            const lon = page.url.searchParams.get("lon");
            map?.setZoom(14);
            map?.setCenter([parseFloat(lon!), parseFloat(lat!)]);
        } else if (settings && settings.mapFocus == "trails") {
            const boundingBox: M.LngLatBoundsLike = [
                [maxBoundingBox.min_lon, maxBoundingBox.max_lat],
                [maxBoundingBox.max_lon, maxBoundingBox.min_lat],
            ];
            map?.fitBounds(boundingBox, { animate: false, padding: 32 });
        } else if (
            settings &&
            settings.mapFocus == "location" &&
            settings.location
        ) {
            map?.setZoom(12);
            map?.setCenter([settings.location.lon, settings.location.lat]);
        } else {
            navigator.geolocation.getCurrentPosition(
                (position) => {
                    const lat = position.coords.latitude;
                    const lon = position.coords.longitude;
                    map?.setZoom(12);
                    map?.setCenter([lon, lat]);
                },
                (error) => {
                    console.error("Error getting user location:", error);
                },
            );
        }
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
                    onclick={() => (showFilter = !showFilter)}
                    ><i class="fa fa-sliders"></i></button
                >
                <button
                    class="btn-icon md:hidden"
                    onclick={() => (showMap = !showMap)}
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
            {#if loading}
                {#each { length: 4 } as _, index}
                    <SkeletonCard></SkeletonCard>
                {/each}
            {:else}
                {#if trails.length == 0}
                    <EmptyStateSearch></EmptyStateSearch>
                {/if}
                {#each trails as trail, i}
                    <a
                        href="map/trail/{trail.id}"
                    >
                        <TrailCard
                            {trail}
                            onmouseenter={() =>
                                handleTrailCardMouseEnter(trail)}
                            onmouseleave={() =>
                                handleTrailCardMouseLeave(trail)}
                        ></TrailCard>
                    </a>
                {/each}
            {/if}
        {/if}
    </div>
    <div
        id="trail-map"
        class:hidden={!showMap && browser && window.innerWidth < 768}
    >
        <MapWithElevationMaplibre
            on:moveend={handleMapMove}
            on:init={handleMapInit}
            {trails}
            showElevation={false}
            showInfoPopup={true}
            activeTrail={-1}
            fitBounds="off"
            minZoom={MIN_ZOOM}
            bind:map
            bind:this={mapWithElevation}
        ></MapWithElevationMaplibre>
    </div>
</main>

<style>
    #trail-map {
        height: calc(100vh - 180px);
    }
    @media only screen and (min-width: 768px) {
        #trail-map,
        #trail-list {
            height: calc(100vh - 124px);
        }
    }
</style>
