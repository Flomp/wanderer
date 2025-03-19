<script lang="ts">
    import type { Category } from "$lib/models/category";
    import type { TrailFilter } from "$lib/models/trail";
    import { country_codes } from "$lib/util/country_code_util";
    import { formatDistance, formatElevation } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import { slide } from "svelte/transition";
    import Datepicker from "../base/datepicker.svelte";
    import DoubleSlider from "../base/double_slider.svelte";
    import MultiSelect from "../base/multi_select.svelte";
    import type { RadioItem } from "../base/radio_group.svelte";
    import RadioGroup from "../base/radio_group.svelte";
    import Search, { type SearchItem } from "../base/search.svelte";
    import type { SelectItem } from "../base/select.svelte";
    import Slider from "../base/slider.svelte";
    import UserSearch from "../user_search.svelte";
    import { pb } from "$lib/pocketbase";
    import { searchLocations } from "$lib/stores/search_store";
    import { getIconForLocation } from "$lib/util/icon_util";
    import { tags_index } from "$lib/stores/tag_store";
    import Combobox, { type ComboboxItem } from "../base/combobox.svelte";
    import { T } from "@threlte/core";

    interface Props {
        categories: Category[];
        filterExpanded?: boolean;
        filter: TrailFilter;
        showTrailSearch?: boolean;
        showCitySearch?: boolean;
        onupdate?: (filter: TrailFilter) => void;
    }

    let {
        categories,
        filterExpanded = $bindable(true),
        filter = $bindable(),
        showTrailSearch = true,
        showCitySearch = true,
        onupdate,
    }: Props = $props();

    let categorySelectItems = $derived(
        categories.map((c) => ({
            value: c.id,
            text: c.name,
        })),
    );

    const radioGroupItems: RadioItem[] = [
        { text: $_("completed"), value: "completed" },
        { text: $_("not-completed"), value: "not_completed" },
        { text: $_("no-preference"), value: "no_preference" },
    ];

    const difficultyItems: SelectItem[] = [
        { text: $_("easy"), value: "easy" },
        { text: $_("moderate"), value: "moderate" },
        { text: $_("difficult"), value: "difficult" },
    ];

    let searchDropdownItems: SearchItem[] = $state([]);

    let citySearchQuery: string = $state("");

    let tagItems: ComboboxItem[] = $state([]);

    async function update() {
        onupdate?.(filter);
    }

    function setCategoryFilter(categories: SelectItem[]) {
        filter.category = categories.map((c) => c.value);

        update();
    }

    function setAuthorFilter(item: SearchItem) {
        filter.author = item.value.id;
        update();
    }

    function setDifficultyFilter(difficulties: SelectItem[]) {
        filter.difficulty = difficulties.map((d) => d.value);
        update();
    }

    function setPublicFilter(e: Event) {
        filter.public = (e.target as HTMLInputElement).checked;
        update();
    }

    function setSharedFilter(e: Event) {
        filter.shared = (e.target as HTMLInputElement).checked;
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
        const r = await searchLocations(q, 5);

        searchDropdownItems = r.map((h) => ({
            text: h.name,
            description: h.description,
            value: h,
            icon: getIconForLocation(h),
        }));
    }

    function handleSearchClick(item: SearchItem) {
        citySearchQuery = item.text;
        filter.near.lat = item.value.lat;
        filter.near.lon = item.value.lon;

        update();
    }

    async function searchTags(q: string) {
        const result = await tags_index(q);
        tagItems = result.items.map((t) => ({ text: t.name, value: t }));
    }

    function getFilterTags(): ComboboxItem[] {
        return filter.tags.map((t) => ({ text: t, value: t }));
    }

    function setFilterTags(tags: ComboboxItem[]) {
        filter.tags = tags.map((t) => t.text);
        update();
    }
</script>

<div class="trail-filter p-8 border border-input-border rounded-xl">
    {#if showTrailSearch}
        <div class="flex gap-2 items-center">
            <div class="basis-full">
                <Search
                    bind:value={filter.q}
                    onupdate={update}
                    placeholder="{$_('search-trails')}..."
                ></Search>
            </div>
            <button
                aria-label="Toggle filter"
                class="btn-icon md:hidden"
                onclick={() => (filterExpanded = !filterExpanded)}
                ><i class="fa fa-sliders"></i></button
            >
        </div>
    {/if}

    {#if filterExpanded}
        <div in:slide out:slide>
            {#if showTrailSearch}
                <hr class="my-4 border-separator" />
            {/if}
            <MultiSelect
                onchange={(value) => setCategoryFilter(value)}
                value={categorySelectItems.filter((i) =>
                    filter.category.includes(i.value),
                )}
                label={$_("categories")}
                items={categorySelectItems}
                placeholder={`${$_("filter-categories")}...`}
            ></MultiSelect>
            <hr class="my-4 border-separator" />
            <Combobox
                bind:value={getFilterTags, setFilterTags}
                onupdate={searchTags}
                placeholder={`${$_("filter-tags")}...`}
                items={tagItems}
                label={$_("tags")}
                multiple
                chips
            ></Combobox>
            <hr class="my-4 border-separator" />

            {#if pb.authStore.record}
                <UserSearch
                    onclick={(item) => setAuthorFilter(item)}
                    onclear={() => {
                        filter.author = "";
                        update();
                    }}
                    clearAfterSelect={false}
                    label={$_("author")}
                ></UserSearch>
                <div class="flex items-center my-4">
                    <input
                        id="public-checkbox"
                        type="checkbox"
                        checked={filter.public}
                        class="w-4 h-4 bg-input-background accent-primary border-input-border focus:ring-input-ring focus:ring-2"
                        onchange={setPublicFilter}
                    />
                    <label for="public-checkbox" class="ms-2 text-sm"
                        >{$_("public")}</label
                    >
                </div>
                <div class="flex items-center my-4">
                    <input
                        id="shared-checkbox"
                        type="checkbox"
                        checked={filter.shared}
                        class="w-4 h-4 bg-input-background accent-primary border-input-border focus:ring-input-ring focus:ring-2"
                        onchange={setSharedFilter}
                    />
                    <label for="shared-checkbox" class="ms-2 text-sm"
                        >{$_("shared")}</label
                    >
                </div>
                <hr class="my-4 border-separator" />
            {/if}
            <MultiSelect
                onchange={(value) => setDifficultyFilter(value)}
                label={$_("difficulty")}
                items={difficultyItems}
                placeholder={`${$_("filter-difficulty")}...`}
            ></MultiSelect>
            <hr class="my-4 border-separator" />
            {#if showCitySearch}
                <div class="mb-8">
                    <Search
                        items={searchDropdownItems}
                        label={$_("near")}
                        placeholder="{$_('search-places')}..."
                        clearAfterSelect={false}
                        bind:value={citySearchQuery}
                        onupdate={(q) => searchCities(q)}
                        onclick={(item) => handleSearchClick(item)}
                    ></Search>
                </div>
                <Slider
                    maxValue={10000}
                    bind:currentValue={filter.near.radius}
                    onset={() => update()}
                ></Slider>
                <p>
                    <span class="text-gray-500 text-sm">{$_("radius")}:</span>
                    {formatDistance(filter.near.radius)}
                </p>
                <hr class="my-4 border-separator" />
            {/if}
            <p class="text-sm font-medium pb-4">{$_("distance")}</p>
            <DoubleSlider
                minValue={filter.distanceMin}
                maxValue={filter.distanceLimit}
                bind:currentMin={filter.distanceMin}
                bind:currentMax={filter.distanceMax}
                onset={() => update()}
            ></DoubleSlider>
            <div class="flex justify-between">
                <span>{formatDistance(filter.distanceMin)}</span>
                <span
                    >{formatDistance(filter.distanceMax)}{filter.distanceMax ==
                    filter.distanceLimit
                        ? "+"
                        : ""}</span
                >
            </div>
            <hr class="my-4 border-separator" />
            <p class="text-sm font-medium pb-4">{$_("elevation-gain")}</p>
            <DoubleSlider
                minValue={filter.elevationGainMin}
                maxValue={filter.elevationGainLimit}
                bind:currentMin={filter.elevationGainMin}
                bind:currentMax={filter.elevationGainMax}
                onset={() => update()}
            ></DoubleSlider>
            <div class="flex justify-between">
                <span>{formatElevation(filter.elevationGainMin)}</span>
                <span
                    >{formatElevation(
                        filter.elevationGainMax,
                    )}{filter.elevationGainMax == filter.elevationGainLimit
                        ? "+"
                        : ""}</span
                >
            </div>
            <hr class="my-4 border-separator" />
            <p class="text-sm font-medium pb-4">{$_("elevation-loss")}</p>
            <DoubleSlider
                minValue={filter.elevationLossMin}
                maxValue={filter.elevationLossLimit}
                bind:currentMin={filter.elevationLossMin}
                bind:currentMax={filter.elevationLossMax}
                onset={() => update()}
            ></DoubleSlider>
            <div class="flex justify-between">
                <span>{formatElevation(filter.elevationLossMin)}</span>
                <span
                    >{formatElevation(
                        filter.elevationLossMax,
                    )}{filter.elevationLossMax == filter.elevationLossLimit
                        ? "+"
                        : ""}</span
                >
            </div>
            <hr class="my-4 border-separator" />

            <div class="space-y-2">
                <Datepicker
                    name="startDate"
                    label={$_("after")}
                    bind:value={filter.startDate}
                    onchange={update}
                ></Datepicker>
                <Datepicker
                    name="endDate"
                    label={$_("before")}
                    bind:value={filter.endDate}
                    onchange={update}
                ></Datepicker>
            </div>

            <hr class="my-4 border-separator" />
            <p class="text-sm font-medium pb-4">{$_("completion-status")}</p>
            <RadioGroup
                name="completed"
                items={radioGroupItems}
                selected={2}
                onchange={(item) => setCompletedFilter(item)}
            ></RadioGroup>
        </div>
    {/if}
</div>
