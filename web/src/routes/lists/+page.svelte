<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import type { DropdownItem } from "$lib/components/base/dropdown.svelte";
    import ConfirmModal from "$lib/components/confirm_modal.svelte";
    import ListCard from "$lib/components/list/list_card.svelte";
    import ListModal from "$lib/components/list/list_modal.svelte";
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
    import type {
        Map
    } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";

    import { tick } from "svelte";
    import { _ } from "svelte-i18n";

    let openListModal: () => void;
    let openConfirmModal: () => void;

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

        if (item.value == "edit") {
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
<main
    class="grid grid-cols-1 md:grid-cols-[430px_1fr] gap-4 lg:gap-4 mx-4"
    style="height: calc(100vh - 124px)"
>
    <ul
        class="list-list mx-4 md:mx-auto rounded-xl border border-input-border max-w-full overflow-y-scroll"
    >
        <div class="flex gap-x-4 items-center p-4 top-0 sticky bg-background z-50">
            <button
                class="flex w-full items-center gap-4 hover:bg-menu-item-background-hover transition-colors rounded-xl p-4 cursor-pointer"
                on:click={() => goto("/lists/edit/new")}
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

        <hr class="border-separator mb-2" />
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
    <div class:hidden={!showMap}>
        <MapWithElevationMultiple trails={$list.expand?.trails ?? []} bind:map options={{itinerary: true, flyToBounds: true}}
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
</main>

<style>
    #map {
        min-height: 600px;
    }
</style>
