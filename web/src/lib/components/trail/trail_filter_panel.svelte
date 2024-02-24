<script lang="ts">
    import { ms } from "$lib/meilisearch";
    import type { Category } from "$lib/models/category";
    import type { TrailFilter } from "$lib/models/trail";
    import { country_codes } from "$lib/util/country_code_util";
    import { formatMeters } from "$lib/util/format_util";
    import { createEventDispatcher } from "svelte";
    import DoubleSlider from "../base/double_slider.svelte";
    import type { RadioItem } from "../base/radio_group.svelte";
    import RadioGroup from "../base/radio_group.svelte";
    import Search, { type SearchItem } from "../base/search.svelte";
    import Slider from "../base/slider.svelte";
    import { slide } from "svelte/transition";

    export let categories: Category[];
    export let filterExpanded: boolean = true;
    export let filter: TrailFilter = {
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
    export let showTrailSearch: boolean = true;
    export let showCitySearch: boolean = true;

    const dispatch = createEventDispatcher();

    const radioGroupItems: RadioItem[] = [
        { text: "Completed", value: "completed" },
        { text: "Not completed", value: "not_completed" },
        { text: "No preference", value: "no_preference" },
    ];

    let searchDropdownItems: SearchItem[] = [];

    let citySearchQuery: string = "";

    async function update() {
        dispatch("update", filter);
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

        update();
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

        update();
    }

    async function searchCities(q: string) {
        if (q.length == 0) {
            filter.near.lat = undefined;
            filter.near.lon = undefined;
            update();

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

        update();
    }
</script>

<div class="trail-filter p-8 border border-input-border rounded-xl">
    {#if showTrailSearch}
        <div class="flex gap-2 items-center">
            <div class="basis-full">
                <Search
                    bind:value={filter.q}
                    on:update={update}
                    placeholder="Search trails..."
                ></Search>
            </div>
            <button
                class="btn-icon md:hidden"
                on:click={() => (filterExpanded = !filterExpanded)}
                ><i class="fa fa-sliders"></i></button
            >
        </div>
    {/if}

    {#if filterExpanded}
        <div in:slide out:slide>
            {#if showTrailSearch}
                <hr class="my-4 border-separator" />
            {/if}
            <p class="text-sm font-medium pb-4">Category</p>
            {#each categories as category, i}
                <div class="flex items-center mb-4">
                    <input
                        id="{category.name}-checkbox"
                        type="checkbox"
                        checked={filter.category.includes(category.id)}
                        value={category.id}
                        class="w-4 h-4 bg-input-background accent-primary border-input-border focus:ring-input-ring focus:ring-2"
                        on:change={() => setCategoryFilter(category)}
                    />
                    <label for="{category.name}-checkbox" class="ms-2 text-sm"
                        >{category.name}</label
                    >
                </div>
            {/each}
            <hr class="my-4 border-separator" />
            {#if showCitySearch}
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
                    on:set={() => update()}
                ></Slider>
                <p>
                    <span class="text-gray-500 text-sm">Radius:</span>
                    {formatMeters(filter.near.radius)}
                </p>
                <hr class="my-4 border-separator" />
            {/if}
            <p class="text-sm font-medium pb-4">Distance</p>
            <DoubleSlider
                minValue={filter.distanceMin}
                maxValue={filter.distanceMax}
                bind:currentMin={filter.distanceMin}
                bind:currentMax={filter.distanceMax}
                on:set={() => update()}
            ></DoubleSlider>
            <div class="flex justify-between">
                <span>{formatMeters(filter.distanceMin)}</span>
                <span>{formatMeters(filter.distanceMax)}</span>
            </div>
            <hr class="my-4 border-separator" />
            <p class="text-sm font-medium pb-4">Elevation Gain</p>
            <DoubleSlider
                minValue={filter.elevationGainMin}
                maxValue={filter.elevationGainMax}
                bind:currentMin={filter.elevationGainMin}
                bind:currentMax={filter.elevationGainMax}
                on:set={() => update()}
            ></DoubleSlider>
            <div class="flex justify-between">
                <span>{formatMeters(filter.elevationGainMin)}</span>
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
        </div>
    {/if}
</div>
