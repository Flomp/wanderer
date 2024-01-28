<script lang="ts">
    import CategoryCard from "$lib/components/category_card.svelte";
    import TrailCard from "$lib/components/trail_card.svelte";
    import { categories } from "$lib/stores/category_store";
    import { trails } from "$lib/stores/trail_store";
</script>

<section class="hero flex justify-center items-center" style="height: 50vh">
    <div class="relative text-gray-600">
        <span class="absolute inset-y-0 left-0 flex items-center pl-4">
            <i class="fa fa-search text-2xl"></i>
        </span>
        <input
            type="search"
            name="q"
            class="py-4 rounded-2xl pl-14 focus:outline-none text-2xl text-primary"
            placeholder="Search trails..."
            autocomplete="off"
        />
    </div>
</section>
<section class="max-w-7xl mx-auto mt-8 px-8 xl:px-0">
    <h2 class="text-5xl md:text-6xl font-bold text-primary">Your trails</h2>
    <div
        id="trails"
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 justify-items-center gap-8 py-8"
    >
        {#each $trails as trail}
            <a href="/trail/{trail.id}"> <TrailCard {trail}></TrailCard></a>
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
