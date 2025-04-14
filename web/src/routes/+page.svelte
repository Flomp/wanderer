<script lang="ts">
    import { goto } from "$app/navigation";
    import emptyStateTrailDark from "$lib/assets/svgs/empty_states/empty_state_trail_dark.svg";
    import emptyStateTrailLight from "$lib/assets/svgs/empty_states/empty_state_trail_light.svg";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import CategoryCard from "$lib/components/category_card.svelte";
    import Scene from "$lib/components/scene.svelte";
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import { defaultTrailSearchAttributes } from "$lib/models/trail.js";
    import { categories } from "$lib/stores/category_store";
    import {
        searchMulti,
        type ListSearchResult,
        type LocationSearchResult,
        type TrailSearchResult,
    } from "$lib/stores/search_store.js";
    import { theme } from "$lib/stores/theme_store";
    import { currentUser } from "$lib/stores/user_store";
    import { getIconForLocation } from "$lib/util/icon_util.js";
    import { Canvas } from "@threlte/core";
    import { _ } from "svelte-i18n";

    let { data } = $props();

    let searchDropdownItems: SearchItem[] = $state([]);

    async function search(q: string) {
        const r = await searchMulti({
            queries: [
                {
                    indexUid: "trails",
                    attributesToRetrieve: defaultTrailSearchAttributes,
                    q: q,
                    limit: 3,
                },
                {
                    indexUid: "lists",
                    q: q,
                    limit: 3,
                },
                {
                    indexUid: "locations",
                    q: q,
                    limit: 5,
                },
            ],
        });

        const trailItems = r[0].hits.map((t: TrailSearchResult) => ({
            text: t.name,
            description: `Trail ${t.location.length ? ", " + t.location : ""}`,
            value: t.id,
            icon: "route",
        }));
        const listItems = r[1].hits.map((t: ListSearchResult) => ({
            text: t.name,
            description: `List, ${t.trails.length} ${$_("trail", { values: { n: t.trails.length } })}`,
            value: t.id,
            icon: "layer-group",
        }));
        const cityItems = r[2].hits.map((c: LocationSearchResult) => ({
            text: c.name,
            description: c.description,
            value: c,
            icon: getIconForLocation(c),
        }));

        searchDropdownItems = [...trailItems, ...listItems, ...cityItems];
    }

    function handleSearchClick(item: SearchItem) {
        if (item.icon == "route") {
            goto(`/trail/view/${item.value}`);
        } else if (item.icon == "layer-group") {
            goto(`/lists?list=${item.value}`);
        } else {
            goto(`/map/?lat=${item.value.lat}&lon=${item.value.lon}`);
        }
    }
</script>

<svelte:head>
    <title>Home | wanderer</title>
</svelte:head>

<section
    class="hero grid grid-cols-1 lg:grid-cols-2 md:px-8 md:gap-8"
    style="height: calc(100vh - 112px)"
>
    <div
        class="flex flex-col justify-center gap-8 max-w-md mx-8 sm:mx-auto mt-0 lg:-mt-24 md:mt-24"
    >
        <h2 class="text-5xl sm:text-6xl md:text-7xl font-bold">
            {$_("welcome_to")} <span class="-tracking-[0.075em]">wanderer</span>
        </h2>
        <h5>
            {$_("hero_section_0_text")}
        </h5>
        <Search
            onupdate={search}
            onclick={handleSearchClick}
            large={true}
            placeholder="{$_('search-for-trails-places')}..."
            items={searchDropdownItems}
        ></Search>
    </div>
    <div class="hidden md:block">
        <Canvas toneMapping={0}>
            <Scene></Scene>
        </Canvas>
    </div>
</section>
<section
    class="max-w-7xl mx-auto mt-8 px-8 xl:px-0 grid grid-cols-1 md:grid-cols-2 items-center gap-x-12"
>
    <div
        id="trails"
        class="flex flex-wrap justify-items-center gap-8 py-8 order-1 md:order-none"
    >
        {#if data.trails.length == 0}
            <img
                style="width: min(450px, 100%)"
                class="rounded-md"
                src={$theme === "light"
                    ? emptyStateTrailLight
                    : emptyStateTrailDark}
                alt="Empty state"
            />
        {:else}
            {#each data.trails as trail}
                <a class="w-full md:max-w-72" href="/trail/view/{trail.id}">
                    <TrailCard {trail} fullWidth></TrailCard></a
                >
            {/each}
        {/if}
    </div>
    <div class="max-w-md md:mx-auto space-y-8">
        {#if data.trails.length == 0}
            <h2 class="text-4xl md:text-5xl font-bold">
                {$_("hero_section_1_heading")}
            </h2>
            <h5>{$_("hero_section_1_text_alternative")}</h5>
            <a
                class="inline-block btn-primary btn-large"
                href="/trail/edit/new"
                role="button">{$_("new-trail")}</a
            >
        {:else}
            <h2 class="text-4xl md:text-5xl font-bold">
                {$currentUser
                    ? $_("trails-for-you")
                    : $_("explore-some-trails")}
            </h2>
            <h5>
                {$_("hero_section_1_text")}
            </h5>
            <a
                class="inline-block btn-primary btn-large"
                href="/trails"
                role="button">{$_("explore")}</a
            >
        {/if}
    </div>
</section>
<section
    class="max-w-7xl mx-auto mt-8 px-8 xl:px-0 grid grid-cols-1 md:grid-cols-2 items-center"
>
    <div class="max-w-md md:mx-auto space-y-8">
        <h2 class="text-4xl md:text-5xl font-bold">{$_("categories")}</h2>
        <h5>
            {$_("hero_section_2_text")}
        </h5>
    </div>
    <div
        id="categories"
        class="grid grid-cols-1 lg:grid-cols-2 justify-items-center gap-8 py-8"
    >
        {#each $categories as category}
            <a href="/trails?category={category.id}">
                <CategoryCard {category}></CategoryCard>
            </a>
        {/each}
    </div>
</section>

<style>
</style>
