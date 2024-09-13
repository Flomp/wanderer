<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import type { DropdownItem } from "$lib/components/base/dropdown.svelte";
    import ConfirmModal from "$lib/components/confirm_modal.svelte";
    import ListCard from "$lib/components/list/list_card.svelte";
    import ListModal from "$lib/components/list/list_modal.svelte";
    import ListShareModal from "$lib/components/list/list_share_modal.svelte";
    import MapWithElevationMultiple from "$lib/components/trail/map_with_elevation_multiple.svelte";
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import { List } from "$lib/models/list";
    import type { TrailFilter } from "$lib/models/trail";
    import {
        list,
        lists,
        lists_create,
        lists_delete,
        lists_index,
        lists_update,
    } from "$lib/stores/list_store";
    import { fetchGPX } from "$lib/stores/trail_store";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import type { Map } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";

    import { tick } from "svelte";
    import { _ } from "svelte-i18n";

    let openListModal: () => void;
    let openConfirmModal: () => void;
    let openShareModal: () => void;

    let filter: TrailFilter = $page.data.filter;
    let listToBeDeleted: List | null = null;

    let map: Map;
    let showMap: boolean = true;

    async function toggleMap() {
        showMap = !showMap;
        if (showMap) {
            setCurrentList($list);
            await tick();
            map.invalidateSize();
        }
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

        if (item.value == "share") {
            list.set(currentList);
            await tick();
            openShareModal();
        } else if (item.value == "edit") {
            goto("/lists/edit/" + currentList.id);
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
        if (item.expand && item.expand.trails.length > 0) {
            for (const trail of item.expand.trails) {
                const gpxData: string = await fetchGPX(trail);
                trail.expand.gpx_data = gpxData;
            }
        }
        list.set(item);
    }
</script>

<svelte:head>
    <title>{$_("list", { values: { n: 2 } })} | wanderer</title>
</svelte:head>
<main class="grid grid-cols-1 md:grid-cols-[430px_1fr] gap-4 lg:gap-4 mx-4">
    <ul
        class="list-list mx-4 md:mx-auto rounded-xl border border-input-border max-h-full"
    >
        <div
            class="flex gap-x-4 items-center justify-between p-4 bg-background z-50 rounded-xl"
        >
            <a class="btn-primary btn-large text-center" href="/lists/edit/new"
                ><i class="fa fa-plus mr-2"></i>{$_("new-list")}</a
            >
            <button type="button" class="btn-icon" on:click={toggleMap}
                ><i class="fa-regular fa-{showMap ? 'rectangle-list' : 'map'}"
                ></i></button
            >
        </div>

        <hr class="border-separator mb-2" />

        <div class="px-4">
            {#each $lists as item, i}
                <li
                    class="list-list-item my-1"
                    on:click={() => setCurrentList(item)}
                    role="presentation"
                >
                    <ListCard
                        list={item}
                        on:change={(e) => handleDropdownClick(e, item)}
                        active={item.id === $list.id}
                    ></ListCard>
                </li>
            {/each}
        </div>
    </ul>
    <div id="trail-map" class="sticky top-[62px]" class:hidden={!showMap}>
        <MapWithElevationMultiple
            trails={$list.expand?.trails ?? []}
            bind:map
            options={{ itinerary: true, flyToBounds: true }}
        ></MapWithElevationMultiple>
    </div>
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
    <ListShareModal list={$list} bind:openShareModal></ListShareModal>
</main>

<style>
    @media only screen and (min-width: 768px) {
        #trail-map {
            height: calc(100vh - 124px);
        }
    }
</style>
