<script lang="ts">
    import { goto } from "$app/navigation";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import CategoryCard from "$lib/components/category_card.svelte";
    import Scene from "$lib/components/scene.svelte";
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import { ms } from "$lib/meilisearch";
    import { categories } from "$lib/stores/category_store";
    import { trails } from "$lib/stores/trail_store";
    import { currentUser } from "$lib/stores/user_store";
    import { country_codes } from "$lib/util/country_code_util";
    import { Canvas } from "@threlte/core";

    let searchDropdownItems: SearchItem[] = [];

    async function search(q: string) {
        const response = await ms.multiSearch({
            queries: [
                {
                    indexUid: "trails",
                    q: q,
                    limit: 3,
                },
                {
                    indexUid: "cities500",
                    q: q,
                    limit: 3,
                },
            ],
        });

        const trailItems = response.results[0].hits.map((t) => ({
            text: t.name,
            description: `Trail | ${t.location}`,
            value: t.id,
            icon: "route",
        }));
        const cityItems = response.results[1].hits.map((t) => ({
            text: t.name,
            description: `City | ${
                country_codes[t["country code"] as keyof typeof country_codes]
            }`,
            value: t,
            icon: "city",
        }));

        searchDropdownItems = [...trailItems, ...cityItems];
    }

    function handleSearchClick(item: SearchItem) {
        if (item.icon == "city") {
            goto(`/map/?lat=${item.value._geo.lat}&lon=${item.value._geo.lng}`);
        }
        if (item.icon == "route") {
            goto(`/trail/view/${item.value}`);
        }
    }
</script>

<section
    class="hero grid grid-cols-1 md:grid-cols-2 md:px-8 md:gap-8"
    style="height: calc(100vh - 112px)"
>
    <div
        class="flex flex-col justify-center gap-8 max-w-md mx-8 md:mx-auto lg:-mt-24"
    >
        <h2 class="text-7xl font-bold">
            Welcome to <span class="-tracking-[0.075em]">wanderer</span>
        </h2>
        <h5>
            Explore exciting trails, save your favorites, and experience the
            beauty of nature. Find your next adventure!
        </h5>
        <Search
            on:update={(e) => search(e.detail)}
            on:click={(e) => handleSearchClick(e.detail)}
            large={true}
            placeholder="Search for trails, places..."
            items={searchDropdownItems}
        ></Search>
    </div>
    <div class="hidden md:block">
        <Canvas>
            <Scene></Scene>
        </Canvas>
    </div>
</section>
<section
    class="max-w-7xl mx-auto mt-8 px-8 xl:px-0 grid grid-cols-1 md:grid-cols-2 items-center"
>
    <div
        id="trails"
        class="flex flex-wrap justify-items-center gap-8 py-8 order-1 md:order-none"
    >
        {#each $trails as trail}
            <a href="/trail/view/{trail.id}">
                <TrailCard {trail}></TrailCard></a
            >
        {/each}
    </div>
    <div class="max-w-md md:mx-auto space-y-8">
        <h2 class="text-4xl md:text-5xl font-bold">
            {$currentUser ? "Trails for you" : "Explore some trails"}
        </h2>
        <h5>
            Here are some trails you might like. Or you can just go to the full
            list right now.
        </h5>
        <a
            class="inline-block btn-primary btn-large"
            href="/trails"
            role="button">Explore</a
        >
    </div>
</section>
<section
    class="max-w-7xl mx-auto mt-8 px-8 xl:px-0 grid grid-cols-1 md:grid-cols-2 items-center"
>
    <div class="max-w-md md:mx-auto space-y-8">
        <h2 class="text-4xl md:text-5xl font-bold">Categories</h2>
        <h5>
            Did you know? You cannot only save you hiking trails. There are many
            categories for all your adventures.
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
