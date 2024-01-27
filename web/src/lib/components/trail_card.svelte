<script lang="ts">
    import type { Trail } from "$lib/models/trail";

    export let trail: Trail;
</script>

<div
    class="trail-card rounded-2xl shadow-md w-72 overflow-hidden cursor-pointer"
>
    <img class="w-full h-48" src={trail.img} alt="" />
    <div class="p-4">
        <div>
            <h4 class="font-semibold text-lg">{trail.name}</h4>
            <h5><i class="fa fa-location-dot mr-3"></i>{trail.location}</h5>
        </div>
        <div class="flex mt-2 gap-4">
            <span><i class="fa fa-left-right mr-2"></i>{trail.distance}m</span>
            <span><i class="fa fa-up-down mr-2"></i>{trail.height}m</span>
        </div>
    </div>
</div>

<style>
    .trail-card img {
        object-fit: cover;
        transition: 0.25s ease;
    }

    .trail-card:hover img {
        scale: 1.075;
    }
</style>
