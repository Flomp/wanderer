<script lang="ts">
    import type { Trail, TrailFilter } from "$lib/models/trail";
    import { _ } from "svelte-i18n";
    import Pagination from "../base/pagination.svelte";
    import Select, { type SelectItem } from "../base/select.svelte";
    import EmptyStateSearch from "../empty_states/empty_state_search.svelte";
    import TrailCard from "./trail_card.svelte";
    import TrailListItem from "./trail_list_item.svelte";
    import TrailTable from "./trail_table.svelte";
    import SkeletonCard from "../base/skeleton_card.svelte";
    import SkeletonListItem from "../base/skeleton_list_item.svelte";
    import { onMount } from "svelte";

    interface Props {
        filter?: TrailFilter | null;
        trails: Trail[];
        pagination?: { page: number; totalPages: number };
        loading?: boolean;
        fullWidthCards?: boolean;
        onupdate?: (filter: TrailFilter | null) => void;
        onpagination?: (page: number) => void;
    }

    let {
        filter = $bindable(null),
        trails,
        pagination = {
            page: 1,
            totalPages: 1,
        },
        loading = false,
        fullWidthCards = false,
        onupdate,
        onpagination,
    }: Props = $props();

    const displayOptions: SelectItem[] = [
        { text: $_("card", { values: { n: 2 } }), value: "cards" },
        { text: $_("list", { values: { n: 1 } }), value: "list" },
        { text: $_("table"), value: "table" },
    ];

    let selectedDisplayOption = $state(displayOptions[0].value);

    const sortOptions: SelectItem[] = [
        { text: $_("name"), value: "name" },
        { text: $_("distance"), value: "distance" },
        { text: $_("duration"), value: "duration" },
        { text: $_("difficulty"), value: "difficulty" },
        { text: $_("elevation-gain"), value: "elevation_gain" },
        { text: $_("elevation-loss"), value: "elevation_loss" },
        { text: $_("creation-date"), value: "created" },
        { text: $_("date"), value: "date" },
    ];

    onMount(() => {
        const storedDisplayOption = localStorage.getItem("displayOption");
        const storedSort = localStorage.getItem("sort");
        const storedSortOrder = localStorage.getItem("sort_order");

        if (storedDisplayOption) {
            selectedDisplayOption = storedDisplayOption;
        }
        if (filter) {
            filter.sort =
                (storedSort as typeof filter.sort | null) ?? filter.sort;
        }
        if (filter) {
            filter.sortOrder =
                (storedSortOrder as typeof filter.sortOrder | null) ??
                filter.sortOrder;
        }
        onupdate?.(filter);
    });

    function setDisplayOption() {
        localStorage.setItem("displayOption", selectedDisplayOption);
    }

    function setSort() {
        if (!filter) {
            return;
        }
        localStorage.setItem("sort", filter.sort);
        onupdate?.(filter);
    }

    function setSortOrder() {
        if (!filter) {
            return;
        }
        if (filter.sortOrder === "+") {
            filter.sortOrder = "-";
        } else {
            filter.sortOrder = "+";
        }
        localStorage.setItem("sort_order", filter.sortOrder);
        onupdate?.(filter);
    }

    function handleSortUpdate(sort: any) {        
        if (!filter) {
            return;
        }
        if (sort === filter.sort) {
            setSortOrder();
        } else {
            filter.sort = sort;
            filter.sortOrder = "+";
            setSort();
        }
    }
</script>

<div class="min-w-0">
    <div class="flex items-end flex-wrap lg:flex-nowrap gap-x-6 gap-y-2 mx-4">
        <div class="basis-full order-1 md:order-none">
            <Pagination
                page={pagination.page}
                totalPages={pagination.totalPages}
                {onpagination}
            ></Pagination>
        </div>
        {#if filter}
            <div class="shrink-0">
                {#if selectedDisplayOption !== "table"}
                    <p class="text-sm text-gray-500 pb-2">{$_("sort")}</p>
                    <div class="flex items-center gap-2">
                        <Select
                            bind:value={filter.sort}
                            items={sortOptions}
                            onchange={setSort}
                        ></Select>
                        <button
                            aria-label="Change sort order"
                            id="sort-order-btn"
                            class="btn-icon"
                            class:rotated={filter.sortOrder == "-"}
                            onclick={() => setSortOrder()}
                            ><i class="fa fa-arrow-up"></i></button
                        >
                    </div>
                {/if}
            </div>
        {/if}
        <div class="shrink-0">
            <p class="text-sm text-gray-500 pb-2">{$_("display-as")}</p>

            <Select
                bind:value={selectedDisplayOption}
                items={displayOptions}
                onchange={() => setDisplayOption()}
            ></Select>
        </div>
    </div>

    <div id="trails" class="flex items-start flex-wrap gap-8 py-8 max-w-full">
        {#if loading}
            {#if selectedDisplayOption === "table"}
                <TrailTable trails={null} tableHeader={sortOptions}
                ></TrailTable>
            {:else}
                {#each { length: 12 } as _, index}
                    {#if selectedDisplayOption === "cards"}
                        <div class="flex-1">
                            <SkeletonCard></SkeletonCard>
                        </div>
                    {:else if selectedDisplayOption === "list"}
                        <SkeletonListItem></SkeletonListItem>
                    {/if}
                {/each}
            {/if}
        {:else}
            {#if trails.length == 0}
                <div class="flex flex-col basis-full items-center">
                    <EmptyStateSearch width={356}></EmptyStateSearch>
                </div>
            {/if}
            {#if selectedDisplayOption === "table"}
                <TrailTable
                    {trails}
                    tableHeader={sortOptions.filter(
                        (option) => option.value !== "elevation_loss",
                    )}
                    {filter}
                    onsort={handleSortUpdate}
                ></TrailTable>
            {:else}
                {#each trails as trail}
                    <a
                        class="max-w-full flex-1"
                        class:basis-full={selectedDisplayOption === "list"}
                        href="/trail/view/{trail.id}"
                    >
                        {#if selectedDisplayOption === "cards"}
                            <TrailCard fullWidth={fullWidthCards} {trail}
                            ></TrailCard>
                        {:else}
                            <TrailListItem {trail}></TrailListItem>
                        {/if}
                    </a>
                {/each}
            {/if}
        {/if}
    </div>
    <Pagination
        page={pagination.page}
        totalPages={pagination.totalPages}
        {onpagination}
    ></Pagination>
</div>

<style>
    #sort-order-btn {
        transition: transform 0.5s ease;
    }
    :global(.rotated) {
        transform: rotate(180deg);
    }
</style>
