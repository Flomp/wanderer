<script lang="ts">
    import CategoryCard from "$lib/components/category_card.svelte";
    import TrailCard from "$lib/components/trail_card.svelte";
    import type { Category } from "$lib/models/category";
    import type { Trail } from "$lib/models/trail";

    const categories: Category[] = [
        {
            name: "Hiking",
            img: "https://images.unsplash.com/photo-1551632811-561732d1e306?q=80&w=3270&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
        },
        {
            name: "Walking",
            img: "https://images.unsplash.com/photo-1606474226448-4aa808468efc?q=80&w=3380&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
        },
        {
            name: "Biking",
            img: "https://images.unsplash.com/photo-1629626495452-37bfe96e142b?q=80&w=3387&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
        },
        {
            name: "Climbing",
            img: "https://images.unsplash.com/photo-1605548109944-9040d0972bf5?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8MTF8fGNsaW1iaW5nfGVufDB8fDB8fHww",
        },
        {
            name: "Skiing",
            img: "https://images.unsplash.com/photo-1616429553002-faf23468952d?q=80&w=3270&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
        },
    ];

    const trails: Trail[] = [
        {
            name: "Breitenstein",
            location: "Fischbachau",
            distance: 4000,
            height: 1400,
            img: "https://images.unsplash.com/photo-1570641963303-92ce4845ed4c?q=80&w=3387&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
            category: categories[0],
            description: ""
        },
    ];
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
        {#each trails as trail,i}
            <a href="/trail/{i}"> <TrailCard {trail}></TrailCard></a>
        {/each}
    </div>
</section>
<section class="max-w-7xl mx-auto mt-8 px-8 xl:px-0">
    <h2 class="text-5xl md:text-6xl font-bold text-primary">By category</h2>
    <div
        id="categories"
        class="grid grid-cols- sm:grid-cols-2 lg:grid-cols-4 justify-items-center gap-8 py-8"
    >
        {#each categories as category}
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
