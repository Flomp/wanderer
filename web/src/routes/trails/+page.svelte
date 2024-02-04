<script lang="ts">
    import DoubleSlider from "$lib/components/base/double_slider.svelte";
    import RadioGroup, {
        type RadioItem,
    } from "$lib/components/base/radio_group.svelte";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import Slider from "$lib/components/base/slider.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import { ms } from "$lib/meilisearch";
    import type { TrailFilter } from "$lib/models/trail";
    import { categories } from "$lib/stores/category_store";
    import { trails } from "$lib/stores/trail_store";
    import { country_codes } from "$lib/util/country_code_util";
    import { formatMeters } from "$lib/util/format_util";

    $: maxDistance = Math.max(...$trails.map((t) => t.distance ?? 0));
    $: maxElevationGain = Math.max(
        ...$trails.map((t) => t.elevation_gain ?? 0),
    );

    const filter: TrailFilter = {
        q: "",
        category: [],
        near: {
            distance: 2000,
        },
        distanceMin: 0,
        distanceMax: maxDistance,
        eleavationGainMin: 0,
        elevationGainMax: maxElevationGain,
    };

    const radioGroupItems: RadioItem[] = [
        { text: "Completed", value: "completed" },
        { text: "Not completed", value: "not_completed" },
        { text: "No preference", value: "no_preference" },
    ];

    let searchDropdownItems: SearchItem[] = [];

    let citySearchQuery: string = "";

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
    }

    async function searchCities(q: string) {        
        if (q.length == 0) {
            filter.near.lat = undefined;
            filter.near.lon = undefined;
            console.log(filter.near);
            
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
    }
</script>

<main class="grid grid-cols-1 md:grid-cols-[300px_1fr] gap-8 max-w-7xl mx-6 md:mx-auto">
    <div class="trail-filters p-8 border rounded-xl">
        <TextField placeholder="Search..."></TextField>
        <hr class="my-4" />
        <p class="text-sm font-medium pb-4">Category</p>
        {#each $categories as category}
            <div class="flex items-center mb-4">
                <input
                    id="{category.name}-checkbox"
                    type="checkbox"
                    value=""
                    class="w-4 h-4 text-primary bg-gray-100 border-gray-300 focus:ring-gray-400 focus:ring-2"
                />
                <label
                    for="{category.name}-checkbox"
                    class="ms-2 text-sm text-gray-900 dark:text-gray-300"
                    >{category.name}</label
                >
            </div>
        {/each}
        <hr class="my-4" />
        <p class="text-sm font-medium pb-4">Near</p>
        <div class="mb-8">
            <Search
                items={searchDropdownItems}
                bind:value={citySearchQuery}
                on:update={(e) => searchCities(e.detail)}
                on:click={(e) => handleSearchClick(e.detail)}
            ></Search>
        </div>
        <Slider maxValue={10000} bind:currentValue={filter.near.distance}
        ></Slider>
        <p>
            <span class="text-gray-500 text-sm">Radius:</span>
            {formatMeters(filter.near.distance)}
        </p>
        <hr class="my-4" />
        <p class="text-sm font-medium pb-4">Distance</p>
        <DoubleSlider
            maxValue={maxDistance}
            bind:currentMin={filter.distanceMin}
            bind:currentMax={filter.distanceMax}
        ></DoubleSlider>
        <div class="flex justify-between">
            <span>{formatMeters(filter.distanceMin)}</span>
            <span>{formatMeters(filter.distanceMax)}</span>
        </div>
        <hr class="my-4" />
        <p class="text-sm font-medium pb-4">Elevation Gain</p>
        <DoubleSlider
            maxValue={maxElevationGain}
            bind:currentMin={filter.eleavationGainMin}
            bind:currentMax={filter.elevationGainMax}
        ></DoubleSlider>
        <div class="flex justify-between">
            <span>{formatMeters(filter.eleavationGainMin)}</span>
            <span>{formatMeters(filter.elevationGainMax)}</span>
        </div>
        <hr class="my-4" />
        <p class="text-sm font-medium pb-4">Completed</p>
        <RadioGroup
            name="completed"
            items={radioGroupItems}
            selected={2}
            on:change={(e) => setCompletedFilter(e.detail)}
        ></RadioGroup>
    </div>
    <div id="trails" class="flex items-start flex-wrap gap-8 py-8">
        {#each $trails as trail}
            <a href="/trail/view/{trail.id}">
                <TrailCard {trail}></TrailCard></a
            >
        {/each}
    </div>
</main>
