<script lang="ts">
    import { goto } from "$app/navigation";
    import Search, {
        type SearchItem,
    } from "$lib/components/base/search.svelte";
    import CategoryCard from "$lib/components/category_card.svelte";
    import TrailCard from "$lib/components/trail/trail_card.svelte";
    import { ms } from "$lib/meilisearch";
    import type { Trail } from "$lib/models/trail";
    import { categories } from "$lib/stores/category_store";
    import {
        trails,
        trails_delete,
        trails_index,
    } from "$lib/stores/trail_store";
    import { currentUser } from "$lib/stores/user_store";
    import { country_codes } from "$lib/util/country_code_util";

    let searchDropdownItems: SearchItem[] = [];

    async function handleDropdownClick(
        currentTrail: Trail,
        item: { text: string; value: any },
    ) {
        if (item.value == "edit") {
            goto(`/trail/edit/${currentTrail.id}`);
        } else if (item.value == "delete") {
            await trails_delete(currentTrail);
            await trails_index();
        }
    }

    async function search(q: string) {
        const response = await ms.multiSearch({
            queries: [
                {
                    indexUid: "trails",
                    q: q,
                    limit: 5,
                },
                {
                    indexUid: "cities500",
                    q: q,
                    limit: 5,
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
            value: t.id,
            icon: "city",
        }));

        searchDropdownItems = [...trailItems, ...cityItems];
    }

    function handleSearchClick(item: SearchItem) {
        if (item.icon == "route") {
            goto(`/trail/view/${item.value}`);
        }
    }
</script>

<section class="hero flex justify-center items-center" style="height: 50vh">
    <Search
        on:update={(e) => search(e.detail)}
        on:click={(e) => handleSearchClick(e.detail)}
        large={true}
        placeholder="Search trails..."
        items={searchDropdownItems}
    ></Search>
</section>
<section class="max-w-7xl mx-auto mt-8 px-8 xl:px-0">
    <h2 class="text-5xl md:text-6xl font-bold text-primary">
        {$currentUser ? "Your" : "Explore"} trails
    </h2>
    <div
        id="trails"
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 justify-items-center gap-8 py-8"
    >
        {#each $trails as trail}
            <a href="/trail/view/{trail.id}">
                <TrailCard
                    {trail}
                    mode="edit"
                    on:change={(e) => handleDropdownClick(trail, e.detail)}
                ></TrailCard></a
            >
        {/each}
    </div>
</section>
<section class="max-w-7xl mx-auto mt-8 px-8 xl:px-0">
    <h2 class="text-5xl md:text-6xl font-bold text-primary">By category</h2>
    <div
        id="categories"
        class="grid grid-cols- sm:grid-cols-2 lg:grid-cols-4 justify-items-center gap-8 py-8"
    >
        {#each $categories as category}
            <CategoryCard {category}></CategoryCard>
        {/each}
    </div>
</section>

<style>
    .hero {
        background-image: url("/imgs/hero.jpg");
        background-position: bottom;
        background-size: cover;
        background-repeat: no-repeat;
    }
</style>
