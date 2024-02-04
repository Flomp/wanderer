<script lang="ts">
    import type { Trail } from "$lib/models/trail";
    import { currentUser } from "$lib/stores/user_store";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
    import Dropdown from "../base/dropdown.svelte";

    export let trail: Trail;
    export let mode: "show" | "edit" = "show";

    const dropdownItems = [
        { text: "Edit", value: "edit" },
        { text: "Delete", value: "delete" },
    ];
</script>

<div class="trail-card rounded-2xl shadow-md sm:w-72 cursor-pointer">
    <div class="w-full min-h-40 max-h-48 overflow-hidden rounded-t-2xl">
        <img src={trail.thumbnail} alt="" />
    </div>
    <div class="p-4">
        <div>
            <div class="flex justify-between items-center">
                <h4 class="font-semibold text-lg">{trail.name}</h4>
                {#if $currentUser && $currentUser.id == trail.author && mode == "edit"}
                    <Dropdown on:change items={dropdownItems}></Dropdown>
                {/if}
            </div>
            <h5><i class="fa fa-location-dot mr-3"></i>{trail.location}</h5>
        </div>
        <div class="flex mt-2 gap-4 text-sm text-gray-500">
            <span
                ><i class="fa fa-left-right mr-2"></i>{formatMeters(
                    trail.distance,
                )}</span
            >
            <span
                ><i class="fa fa-up-down mr-2"></i>{formatMeters(
                    trail.elevation_gain,
                )}</span
            >
            <span
                ><i class="fa fa-clock mr-2"></i>{formatTimeHHMM(
                    trail.duration,
                )}</span
            >
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
