<script lang="ts">
    import type { Trail, TrailFilter } from "$lib/models/trail";
    import { createEventDispatcher, onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import Pagination from "../base/pagination.svelte";
    import Select, { type SelectItem } from "../base/select.svelte";
    import EmptyStateSearch from "../empty_states/empty_state_search.svelte";
    import TrailCard from "./trail_card.svelte";
    import TrailListItem from "./trail_list_item.svelte";

    export let filter: TrailFilter;
    export let trails: Trail[];
    export let pagination: { page: number; totalPages: number } = {
        page: 1,
        totalPages: 1,
    };

    const displayOptions: SelectItem[] = [
        { text: $_("card", { values: { n: 2 } }), value: "cards" },
        { text: $_("list", { values: { n: 1 } }), value: "list" },
    ];

    let selectedDisplayOption = displayOptions[0].value;

    let dispatch = createEventDispatcher();

    const sortOptions: SelectItem[] = [
        { text: $_("alphabetical"), value: "name" },
        { text: $_("creation-date"), value: "created" },
        { text: $_("date"), value: "date" },
        { text: $_("difficulty"), value: "difficulty" },
        { text: $_("distance"), value: "distance" },
        { text: $_("elevation-gain"), value: "elevation_gain" },
    ];

    onMount(() => {
        const storedDisplayOption = localStorage.getItem("displayOption");
        const storedSort= localStorage.getItem("sort");
        const storedSortOrder= localStorage.getItem("sort_order");

        if (storedDisplayOption) {
            selectedDisplayOption = storedDisplayOption;
        }
        if(storedSort) {
            filter.sort = storedSort as typeof filter.sort
        }
        if(storedSortOrder) {
            filter.sortOrder = storedSortOrder as typeof filter.sortOrder
        }
        if(storedSort || storedSortOrder) {                        
            dispatch("update", filter);
        }
    });

    function setDisplayOption() {
        localStorage.setItem("displayOption", selectedDisplayOption);
    }

    function setSort() {
        localStorage.setItem("sort", filter.sort)
        dispatch("update", filter);
    }

    function setSortOrder() {
        if (filter.sortOrder === "+") {
            filter.sortOrder = "-";
        } else {
            filter.sortOrder = "+";
        }
        localStorage.setItem("sort_order", filter.sortOrder)
        dispatch("update", filter);
    }
</script>

<div class="min-w-0">
    <div class="flex items-end flex-wrap md:flex-nowrap gap-x-6 gap-y-2 mx-4">
        <div class="basis-full order-1 md:order-none">
            <Pagination
                page={pagination.page}
                totalPages={pagination.totalPages}
                on:pagination
            ></Pagination>
        </div>
        <div class="shrink-0">
            <p class="text-sm text-gray-500 pb-2">{$_("sort")}</p>
            <div class="flex items-center gap-2">
                <Select
                    bind:value={filter.sort}
                    items={sortOptions}
                    on:change={setSort}
                ></Select>
                <button
                    id="sort-order-btn"
                    class="btn-icon"
                    class:rotated={filter.sortOrder == "-"}
                    on:click={() => setSortOrder()}
                    ><i class="fa fa-arrow-up"></i></button
                >
            </div>
        </div>
        <div class="shrink-0">
            <p class="text-sm text-gray-500 pb-2">{$_("display-as")}</p>

            <Select
                bind:value={selectedDisplayOption}
                items={displayOptions}
                on:change={() => setDisplayOption()}
            ></Select>
        </div>
    </div>

    <div id="trails" class="flex items-start flex-wrap gap-8 py-8 max-w-full">
        {#if trails.length == 0}
            <div class="flex flex-col basis-full items-center">
                <EmptyStateSearch width={356}></EmptyStateSearch>
            </div>
        {/if}
        {#each trails as trail}
            <a
                class="max-w-full"
                class:basis-full={selectedDisplayOption === "list"}
                href="/trail/view/{trail.id}"
                data-sveltekit-preload-data="off"
            >
                {#if selectedDisplayOption === "cards"}
                    <TrailCard {trail}></TrailCard>
                {:else}
                    <TrailListItem {trail}></TrailListItem>
                {/if}
            </a>
        {/each}
    </div>
</div>

<style>
    #sort-order-btn {
        transition: transform 0.5s ease;
    }
    :global(.rotated) {
        transform: rotate(180deg);
    }
</style>
