<script lang="ts">
    import { goto } from "$app/navigation";
    import type { Trail } from "$lib/models/trail";
    import { trails_delete, trails_index } from "$lib/stores/trail_store";
    import { currentUser } from "$lib/stores/user_store";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
    import Dropdown, { type DropdownItem } from "../base/dropdown.svelte";

    export let trail: Trail;
    export let mode: "show" | "edit" = "show";

    const dropdownItems: DropdownItem[] = [
        { text: "Open", value: "open", icon: "up-right-from-square" },
        { text: "Show on map", value: "map", icon: "map" },
        { text: "Edit", value: "edit", icon: "pen" },
    ];

    async function handleDropdownClick(
        currentTrail: Trail,
        item: { text: string; value: any },
    ) {
        if (item.value == "open") {
            goto(`/trail/view/${currentTrail.id}`);
        } else if (item.value == "map") {
            goto(`/map/?lat=${currentTrail.lat}&lon=${currentTrail.lon}`);
        } else if (item.value == "edit") {
            goto(`/trail/edit/${currentTrail.id}`);
        }
    }

    async function deleteTrail() {
        await trails_delete(trail);
        await trails_index();
    }
</script>

<div
    class="trail-card rounded-2xl border sm:w-72 cursor-pointer"
    on:mouseenter
    on:mouseleave
    role="listitem"
>
    <div class="w-full min-h-40 max-h-48 overflow-hidden rounded-t-2xl">
        <img src={trail.thumbnail} alt="" />
    </div>
    <div class="p-4">
        <div>
            <div class="flex justify-between items-center">
                <h4 class="font-semibold text-lg">{trail.name}</h4>
                {#if $currentUser && $currentUser.id == trail.author && mode == "edit"}
                    <Dropdown
                        on:change={(e) => handleDropdownClick(trail, e.detail)}
                        items={dropdownItems}
                    ></Dropdown>
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
