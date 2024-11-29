<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import type { DropdownItem } from "$lib/components/base/dropdown.svelte";
    import ConfirmModal from "$lib/components/confirm_modal.svelte";
    import ListCard from "$lib/components/list/list_card.svelte";
    import ListPanel from "$lib/components/list/list_panel.svelte";
    import ListShareModal from "$lib/components/list/list_share_modal.svelte";
    import MapWithElevationMaplibre from "$lib/components/trail/map_with_elevation_maplibre.svelte";
    import TrailInfoPanel from "$lib/components/trail/trail_info_panel.svelte";
    import TrailList from "$lib/components/trail/trail_list.svelte";
    import { List } from "$lib/models/list";
    import type { Trail } from "$lib/models/trail";
    import {
        lists,
        lists_delete,
        lists_index
    } from "$lib/stores/list_store";
    import { fetchGPX } from "$lib/stores/trail_store";
    import * as M from "maplibre-gl";

    import { onMount, tick } from "svelte";
    import { _ } from "svelte-i18n";

    let openConfirmModal: () => void;
    let openShareModal: () => void;

    let map: M.Map;
    let mapWithElevation: MapWithElevationMaplibre;
    let markers: any[];
    let showMap: boolean = true;

    let selectedList: List | null = null;
    let selectedTrail: Trail | null = null;

    let activeTrailIndex: number = -1;

    onMount(() => {
        if ($page.url.searchParams.get("list")) {
            const listToFocus = $lists.find(
                (l) => l.id == $page.url.searchParams.get("list"),
            );
            if (listToFocus) {
                setCurrentList(listToFocus);
            }
        }
    });

    async function handleDropdownClick(e: CustomEvent<DropdownItem>) {
        const item = e.detail;

        if (!selectedList) {
            return;
        }
        if (item.value == "share") {
            openShareModal();
        } else if (item.value == "edit") {
            goto("/lists/edit/" + selectedList.id);
        } else if (item.value == "delete") {
            openConfirmModal();
        }
    }

    async function deleteList() {
        if (!selectedList) {
            return;
        }
        await lists_delete(selectedList);
        await lists_index();
        selectedList = null;
    }

    async function setCurrentList(item: List) {
        if (
            item.expand &&
            item.expand.trails &&
            item.expand.trails.length > 0
        ) {
            for (const trail of item.expand.trails) {
                const gpxData: string = await fetchGPX(trail);
                trail.expand.gpx_data = gpxData;
            }
        }
        selectedList = item;
    }

    async function back() {
        if (selectedTrail) {
            mapWithElevation.unFocusTrail(selectedTrail);
            selectedTrail = null;
        } else if (selectedList) {
            selectedList = null;
            map.flyTo({
                animate: true,
                zoom: 1,
                center: [0, 0],
            });
        }
    }

    function selectTrail(trail: Trail) {
        selectedTrail = trail;
        mapWithElevation.focusTrail(trail);
        window.scrollTo({ top: 0 });
    }

    function highlightTrail(trail: Trail) {
        mapWithElevation.highlightTrail(trail.id!);
    }

    function unHighlightTrail(trail: Trail) {
        mapWithElevation.unHighlightTrail(trail.id!);
    }
</script>

<svelte:head>
    <title>{$_("list", { values: { n: 2 } })} | wanderer</title>
</svelte:head>
<main class="grid grid-cols-1 md:grid-cols-[430px_1fr] gap-4 lg:gap-4 md:mx-4">
    <div
        class="list-list md:mx-auto rounded-xl border border-input-border max-h-full w-full order-1 md:order-none"
    >
        <div
            class="flex gap-x-4 items-center justify-between p-4 bg-background z-50 rounded-xl"
        >
            <button
                class="btn-icon"
                class:btn-disabled={!selectedList}
                disabled={!selectedList}
                on:click={back}><i class="fa fa-arrow-left"></i></button
            >
            <a class="btn-primary btn-large text-center" href="/lists/edit/new"
                ><i class="fa fa-plus mr-2"></i>{$_("new-list")}</a
            >
            <!-- <button type="button" class="btn-icon" on:click={toggleMap}
                ><i class="fa-regular fa-{showMap ? 'rectangle-list' : 'map'}"
                ></i></button
            > -->
        </div>

        <hr class="border-separator" />

        <div>
            {#if !selectedList}
                <div class="px-4 mt-2">
                    {#each $lists as item, i}
                        <div
                            class="list-list-item my-1"
                            on:click={() => setCurrentList(item)}
                            role="presentation"
                        >
                            <ListCard list={item}></ListCard>
                        </div>
                    {/each}
                </div>
            {:else if selectedList && !selectedTrail}
                <ListPanel
                    list={selectedList}
                    on:mouseenter={(e) => highlightTrail(e.detail.trail)}
                    on:mouseleave={(e) => unHighlightTrail(e.detail.trail)}
                    on:change={(e) => handleDropdownClick(e)}
                    on:click={(e) => selectTrail(e.detail.trail)}
                ></ListPanel>
            {:else if selectedList && selectedTrail}
                <TrailInfoPanel trail={selectedTrail} mode="list" {markers}
                ></TrailInfoPanel>
            {/if}
        </div>
    </div>
    <div id="trail-map" class="md:sticky md:top-[62px]" class:hidden={!showMap}>
        <MapWithElevationMaplibre
            trails={selectedList?.expand?.trails ?? []}
            bind:map
            bind:this={mapWithElevation}
            bind:markers
            on:select={(e) => {
                selectedTrail = e.detail;
            }}
            bind:activeTrail={activeTrailIndex}
            showInfoPopup={true}
        ></MapWithElevationMaplibre>
    </div>
    <div class="min-w-0" class:hidden={showMap}>
        <TrailList trails={selectedList?.expand?.trails ?? []}></TrailList>
    </div>

    <ConfirmModal
        text={$_("delete-list-confirm")}
        bind:openModal={openConfirmModal}
        on:confirm={deleteList}
    ></ConfirmModal>
    {#if selectedList}
        <ListShareModal list={selectedList} bind:openShareModal
        ></ListShareModal>
    {/if}
</main>

<style>
    @media only screen and (min-width: 768px) {
        #trail-map {
            height: calc(100vh - 124px);
        }
    }
</style>
