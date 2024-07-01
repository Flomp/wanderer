<script lang="ts">
    import { page } from "$app/stores";
    import type { DropdownItem } from "$lib/components/base/dropdown.svelte";
    import ConfirmModal from "$lib/components/confirm_modal.svelte";
    import ListCard from "$lib/components/list/list_card.svelte";
    import ListModal from "$lib/components/list/list_modal.svelte";
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import { List } from "$lib/models/list";
    import type { Trail, TrailFilter } from "$lib/models/trail";
    import {
        list,
        lists,
        lists_create,
        lists_delete,
        lists_index,
        lists_update,
    } from "$lib/stores/list_store";
    import { fetchGPX } from "$lib/stores/trail_store";
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
        LatLngBoundsExpression,
        LeafletEvent,
        Map,
        Marker,
    } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";

    import { onMount, tick } from "svelte";
    import { _ } from "svelte-i18n";

    let openListModal: () => void;
    let openConfirmModal: () => void;

    let filter: TrailFilter = $page.data.filter;
    let listToBeDeleted: List | null = null;

    let L: any;
    let map: Map;
    let showMap: boolean = false;
    let gpxLayers: GPX[] = [];

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");

        map = L.map("map").setView([0, 0], 4);
        map.attributionControl.setPrefix(false);

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);
    });

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

    async function toggleMap() {
        showMap = !showMap;
        if (showMap) {
            setCurrentList($list);
            await tick();
            map.invalidateSize();
        }
    }

    function beforeListModalOpen() {
        list.set(new List("", []));
        openListModal();
    }

    async function saveList(e: CustomEvent<{ list: List; avatar?: File }>) {
        const result = e.detail;
        if (result.list.id) {
            await lists_update(result.list, result.avatar);
        } else {
            await lists_create(result.list, result.avatar);
        }
        await lists_index();
    }

    async function handleDropdownClick(
        e: CustomEvent<DropdownItem>,
        currentList: List,
    ) {
        const item = e.detail;

        if (item.value == "edit") {
            list.set(currentList);
            openListModal();
        } else if (item.value == "delete") {
            openConfirmModal();
            listToBeDeleted = currentList;
        }
    }

    async function deleteList() {
        if (!listToBeDeleted) {
            return;
        }
        await lists_delete(listToBeDeleted);
        await lists_index();
        listToBeDeleted = null;
    }

    async function setCurrentList(item: List) {
        for (const layer of gpxLayers) {
            map.removeLayer(layer);
        }

        list.set(item);

        if (item.expand && item.expand.trails.length > 0) {
            let minLat = item.expand.trails[0].lat!;
            let maxLat = item.expand.trails[0].lat!;
            let minLon = item.expand.trails[0].lon!;
            let maxLon = item.expand.trails[0].lon!;

            for (const trail of item.expand.trails) {
                minLat = Math.min(minLat, trail.lat!);
                maxLat = Math.max(maxLat, trail.lat!);
                minLon = Math.min(minLon, trail.lon!);
                maxLon = Math.max(maxLon, trail.lon!);

                const gpxData: string = await fetchGPX(trail);
                trail.expand.gpx_data = gpxData;
                gpxLayers.push(await addGPXLayer(trail));
            }
            const boundingBox: LatLngBoundsExpression = [
                [maxLat, minLon],
                [minLat, maxLon],
            ];
            map.fitBounds(boundingBox);
        } else {
            map.setView([0, 0], 4);
        }
    }
</script>

<svelte:head>
    <title>{$_("list", { values: { n: 2 } })} | wanderer</title>
</svelte:head>
<main
    class="grid grid-cols-1 md:grid-cols-[400px_1fr] gap-4 lg:gap-8 max-w-7xl mx-4 md:mx-auto"
    style="min-height: calc(100vh - 124px)"
>
    <ul
        class="list-list mx-2 md:mx-auto rounded-xl border border-input-border p-4 max-w-full"
    >
        <div class="flex gap-x-4 items-center">
            <button
                class="flex w-full items-center gap-4 hover:bg-menu-item-background-hover transition-colors rounded-xl p-4 cursor-pointer"
                on:click={beforeListModalOpen}
            >
                <i class="fa fa-plus text-xl aspect-square"></i>
                <h5 id="create-list-button" class="text-xl font-semibold">
                    {$_("create-new-list")}
                </h5>
            </button>
            <button type="button" class="btn-icon" on:click={toggleMap}
                ><i class="fa-regular fa-{showMap ? 'rectangle-list' : 'map'}"
                ></i></button
            >
        </div>

        <hr class="border-separator my-2" />
        {#each $lists as item, i}
            <li
                class="list-list-item"
                on:click={() => setCurrentList(item)}
                role="presentation"
            >
                <ListCard
                    list={item}
                    on:change={(e) => handleDropdownClick(e, item)}
                    active={item.id === $list.id}
                ></ListCard>
                {#if i != $lists.length - 1}
                    <hr class="border-separator my-2" />
                {/if}
            </li>
        {/each}
    </ul>
    <div id="map" class="rounded-xl z-0" class:hidden={!showMap}></div>
    <div class="min-w-0" class:hidden={showMap}>
        <TrailList
            bind:filter
            trails={$list.expand?.trails ?? []}
            on:update={async () => await lists_index()}
        ></TrailList>
    </div>

    <ListModal bind:openModal={openListModal} on:save={saveList}></ListModal>
    <ConfirmModal
        text={$_("delete-list-confirm")}
        bind:openModal={openConfirmModal}
        on:confirm={deleteList}
    ></ConfirmModal>
</main>

<style>
    #map {
        min-height: 600px;
    }
</style>
