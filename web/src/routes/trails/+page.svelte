<script lang="ts">
    import DoubleSlider from "$lib/components/base/double_slider.svelte";
    import RadioGroup, {
        type RadioItem,
    } from "$lib/components/base/radio_group.svelte";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import Select, {
        type SelectItem,
    } from "$lib/components/base/select.svelte";
    import Slider from "$lib/components/base/slider.svelte";
    import EmptyStateSearch from "$lib/components/empty_states/empty_state_search.svelte";
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import TrailListItem from "$lib/components/trail/trail_list_item.svelte";
    import { ms } from "$lib/meilisearch";
    import type { Category } from "$lib/models/category";
    import type { TrailFilter } from "$lib/models/trail";
    import { categories } from "$lib/stores/category_store";
    import { trails, trails_search_filter } from "$lib/stores/trail_store";
    import { country_codes } from "$lib/util/country_code_util";
    import { formatMeters } from "$lib/util/format_util";
    import { onMount } from "svelte";

    $: minDistance = Math.floor(
        Math.min(...$trails.map((t) => t.distance ?? 0)),
    );
    $: maxDistance = Math.ceil(
        Math.max(...$trails.map((t) => t.distance ?? 0)),
    );
    $: minElevationGain = Math.floor(
        Math.min(...$trails.map((t) => t.elevation_gain ?? 0)),
    );
    $: maxElevationGain = Math.ceil(
        Math.max(...$trails.map((t) => t.elevation_gain ?? 0)),
    );

    const filter: TrailFilter = {
        q: "",
        category: [],
        near: {
            radius: 2000,
        },
        distanceMin: minDistance,
        distanceMax: maxDistance,
        eleavationGainMin: minElevationGain,
        elevationGainMax: maxElevationGain,
        sort: "created",
        sortOrder: "+",
    };

    const radioGroupItems: RadioItem[] = [
        { text: "Completed", value: "completed" },
        { text: "Not completed", value: "not_completed" },
        { text: "No preference", value: "no_preference" },
    ];

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

    let searchDropdownItems: SearchItem[] = [];

    let citySearchQuery: string = "";

    let filterExpanded: boolean = true;

    onMount(() => {
        const storedDisplayOption = localStorage.getItem("displayOption");
        if (storedDisplayOption) {
            selectedDisplayOption = storedDisplayOption;
        }

        if (window.innerWidth < 768) {
            filterExpanded = false;
        }
    });

    async function searchTrails() {
        await trails_search_filter(filter);
    }

    function setCategoryFilter(category: Category) {
        const categoryIndex = filter.category.findIndex(
            (c) => c == category.id,
        );
        if (categoryIndex !== -1) {
            filter.category.splice(categoryIndex, 1);
        } else {
            filter.category.push(category.id);
        }

        searchTrails();
    }

    function setCompletedFilter(item: RadioItem) {
        switch (item.value) {
            case "no_preference":
                filter.completed = undefined;
                break;
            case "completed":
                filter.completed = true;
                break;
            case "not_completed":
                filter.completed = false;
                break;
            default:
                filter.completed = undefined;
                break;
        }

        searchTrails();
    }

    async function searchCities(q: string) {
        if (q.length == 0) {
            filter.near.lat = undefined;
            filter.near.lon = undefined;
            searchTrails();

            return;
        }
        const result = await ms.index("cities500").search(q, { limit: 5 });
        searchDropdownItems = result.hits.map((h) => ({
            text: h.name,
            description:
                country_codes[h["country code"] as keyof typeof country_codes],
            value: h,
            icon: "city",
        }));
    }

    function handleSearchClick(item: SearchItem) {
        citySearchQuery = item.text;
        filter.near.lat = item.value.lat;
        filter.near.lon = item.value.lon;

        searchTrails();
    }

    function setSort() {
        searchTrails();
    }

    function setSortOrder() {
        if (filter.sortOrder === "+") {
            filter.sortOrder = "-";
        } else {
            filter.sortOrder = "+";
        }

        searchTrails();
    }

    function setDisplayOption() {
        localStorage.setItem("displayOption", selectedDisplayOption);
    }
</script>

<main
    class="grid grid-cols-1 md:grid-cols-[300px_1fr] gap-8 max-w-7xl mx-6 md:mx-auto"
>
    <div class="trail-filters p-8 border border-input-border rounded-xl">
        <div class="flex gap-2 items-center">
            <div class="basis-full">
                <Search
                    bind:value={filter.q}
                    on:update={searchTrails}
                    placeholder="Search trails..."
                ></Search>
            </div>
            <button
                id="sort-order-btn"
                class="btn-icon md:hidden"
                on:click={() => (filterExpanded = !filterExpanded)}
                ><i class="fa fa-sliders"></i></button
            >
        </div>

        {#if filterExpanded}
            <hr class="my-4 border-separator" />
            <p class="text-sm font-medium pb-4">Category</p>
            {#each $categories as category, i}
                <div class="flex items-center mb-4">
                    <input
                        id="{category.name}-checkbox"
                        type="checkbox"
                        value={category.id}
                        class="w-4 h-4 accent-input-background border-input-border focus:ring-input-ring focus:ring-2"
                        on:change={() => setCategoryFilter(category)}
                    />
                    <label
                        for="{category.name}-checkbox"
                        class="ms-2 text-sm"
                        >{category.name}</label
                    >
                </div>
            {/each}
            <hr class="my-4 border-separator" />
            <p class="text-sm font-medium pb-4">Near</p>
            <div class="mb-8">
                <Search
                    items={searchDropdownItems}
                    placeholder="Search cities..."
                    bind:value={citySearchQuery}
                    on:update={(e) => searchCities(e.detail)}
                    on:click={(e) => handleSearchClick(e.detail)}
                ></Search>
            </div>
            <Slider
                maxValue={10000}
                bind:currentValue={filter.near.radius}
                on:set={() => searchTrails()}
            ></Slider>
            <p>
                <span class="text-gray-500 text-sm">Radius:</span>
                {formatMeters(filter.near.radius)}
            </p>
            <hr class="my-4 border-separator" />
            <p class="text-sm font-medium pb-4">Distance</p>
            <DoubleSlider
                minValue={minDistance}
                maxValue={maxDistance}
                bind:currentMin={filter.distanceMin}
                bind:currentMax={filter.distanceMax}
                on:set={() => searchTrails()}
            ></DoubleSlider>
            <div class="flex justify-between">
                <span>{formatMeters(filter.distanceMin)}</span>
                <span>{formatMeters(filter.distanceMax)}</span>
            </div>
            <hr class="my-4 border-separator" />
            <p class="text-sm font-medium pb-4">Elevation Gain</p>
            <DoubleSlider
                minValue={minElevationGain}
                maxValue={maxElevationGain}
                bind:currentMin={filter.eleavationGainMin}
                bind:currentMax={filter.elevationGainMax}
                on:set={() => searchTrails()}
            ></DoubleSlider>
            <div class="flex justify-between">
                <span>{formatMeters(filter.eleavationGainMin)}</span>
                <span>{formatMeters(filter.elevationGainMax)}</span>
            </div>
            <hr class="my-4 border-separator" />
            <p class="text-sm font-medium pb-4">Completed</p>
            <RadioGroup
                name="completed"
                items={radioGroupItems}
                selected={2}
                on:change={(e) => setCompletedFilter(e.detail)}
            ></RadioGroup>
        {/if}
    </div>
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
                    <div class="w-72 md:w-96 my-4">
                        <EmptyStateSearch></EmptyStateSearch>
                    </div>
                    <h3 class="text-xl md:text-3xl font-semibold text-center">
                        No results found
                    </h3>
                </div>
            {/if}
            {#each $trails as trail}
                <a
                    class="max-w-full"
                    class:basis-full={selectedDisplayOption === "list"}
                    href="/trail/view/{trail.id}"
                >
                    {#if selectedDisplayOption === "cards"}
                        <TrailCard {trail} mode="edit"></TrailCard>
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
