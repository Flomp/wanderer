<script lang="ts">
    import Select, {
        type SelectItem,
    } from "$lib/components/base/select.svelte";
    import EmptyStateSearch from "$lib/components/empty_states/empty_state_search.svelte";
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import TrailFilterPanel from "$lib/components/trail/trail_filter_panel.svelte";
    import TrailListItem from "$lib/components/trail/trail_list_item.svelte";
    import type { TrailFilter } from "$lib/models/trail";
    import { categories } from "$lib/stores/category_store";
    import { trails, trails_search_filter } from "$lib/stores/trail_store";
    import { onMount } from "svelte";

    const displayOptions: SelectItem[] = [
        { text: "Cards", value: "cards" },
        { text: "List", value: "list" },
    ];
    let selectedDisplayOption = displayOptions[0].value;

    const sortOptions: SelectItem[] = [
        { text: "Alphabetical", value: "name" },
        { text: "Creation date", value: "created" },
        { text: "Distance", value: "distance" },
        { text: "Elevation gain", value: "elevation_gain" },
    ];

    let filterExpanded: boolean = true;

    export const filter: TrailFilter = {
        q: "",
        category: [],
        near: {
            radius: 2000,
        },
        distanceMin: 0,
        distanceMax: 20000,
        elevationGainMin: 0,
        elevationGainMax: 4000,
        sort: "created",
        sortOrder: "+",
    };

    onMount(() => {
        const storedDisplayOption = localStorage.getItem("displayOption");
        if (storedDisplayOption) {
            selectedDisplayOption = storedDisplayOption;
        }

        if (window.innerWidth < 768) {
            filterExpanded = false;
        }
    });

    function setDisplayOption() {
        localStorage.setItem("displayOption", selectedDisplayOption);
    }

    async function setSort() {
        await trails_search_filter(filter);
    }

    async function setSortOrder() {
        if (filter.sortOrder === "+") {
            filter.sortOrder = "-";
        } else {
            filter.sortOrder = "+";
        }

        await trails_search_filter(filter);
    }

    async function handleFilterUpdate() {
        await trails_search_filter(filter);
    }
</script>

<main
    class="grid grid-cols-1 md:grid-cols-[300px_1fr] gap-8 max-w-7xl mx-6 md:mx-auto"
>
    <TrailFilterPanel
        categories={$categories}
        {filter}
        {filterExpanded}
        on:update={() => handleFilterUpdate()}
    ></TrailFilterPanel>
    <div class="min-w-0">
        <div class="flex items-start gap-8 justify-end mx-4">
            <div>
                <p class="text-sm text-gray-500 pb-2">Sort</p>
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
            <div>
                <p class="text-sm text-gray-500 pb-2">Display as</p>

                <Select
                    bind:value={selectedDisplayOption}
                    items={displayOptions}
                    on:change={() => setDisplayOption()}
                ></Select>
            </div>
        </div>

        <div
            id="trails"
            class="flex items-start flex-wrap gap-8 py-8 max-w-full"
        >
            {#if $trails.length == 0}
                <div class="flex flex-col basis-full items-center">
                    <EmptyStateSearch></EmptyStateSearch>
                </div>
            {/if}
            {#each $trails as trail}
                <a
                    class="max-w-full"
                    class:basis-full={selectedDisplayOption === "list"}
                    href="/trail/view/{trail.id}"
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
</main>

<style>
    #sort-order-btn {
        transition: transform 0.5s ease;
    }
    :global(.rotated) {
        transform: rotate(180deg);
    }
</style>
