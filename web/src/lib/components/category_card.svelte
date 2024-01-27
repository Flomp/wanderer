<script lang="ts">
    import type { Category } from "$lib/models/category";

    export let category: Category;
</script>

<div class="category-card relative rounded-2xl shadow-md w-48 h-48 overflow-hidden cursor-pointer">
    <img class="w-full h-full" src={category.img} alt="" />
    <div class="absolute bottom-0 w-full h-1/2 bg-gradient-to-b from-transparent to-black opacity-50"></div>
    <h5 class="absolute text-white font-bold bottom-4 left-4 text-xl">{category.name}</h5>
</div>

<style>
    .category-card img {
        object-fit: cover;
        transition: 0.25s ease;
    }

    .category-card:hover img {
        scale: 1.075;
    }
</style>
