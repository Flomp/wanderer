<script lang="ts">
    import { goto } from "$app/navigation";
    import type { Trail } from "$lib/models/trail";
    import { trails_delete, trails_index } from "$lib/stores/trail_store";
    import { currentUser } from "$lib/stores/user_store";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
    import Dropdown from "../base/dropdown.svelte";

    export let trail: Trail;

    const dropdownItems = [
        { text: "Edit", value: "edit" },
        { text: "Delete", value: "delete" },
    ];

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
</script>

<li
    class="flex gap-8 p-4 rounded-xl shadow-md cursor-pointer hover:bg-gray-100 transition-colors"
>
    <div class="shrink-0">
        <img
            class="h-28 w-28 object-cover rounded-xl"
            src={trail.thumbnail}
            alt=""
        />
    </div>
    <div class="min-w-0">
        <div class="flex items-center justify-between">
            <h4 class="font-semibold text-lg">{trail.name}</h4>

            {#if $currentUser && $currentUser.id == trail.author}
                <Dropdown
                    on:change={(e) => handleDropdownClick(trail, e.detail)}
                    items={dropdownItems}
                ></Dropdown>
            {/if}
        </div>
        <h5><i class="fa fa-location-dot mr-3"></i>{trail.location}</h5>
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
        <p
            class="mt-3 text-sm whitespace-nowrap min-w-0 max-w-full overflow-hidden text-ellipsis"
        >
            {trail.description}
        </p>
    </div>
</li>
