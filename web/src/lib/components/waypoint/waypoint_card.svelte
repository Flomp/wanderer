<script lang="ts">
    import type { Waypoint } from "$lib/models/waypoint";
    import { _ } from "svelte-i18n";
    import Dropdown from "../base/dropdown.svelte";
    
    export let waypoint: Waypoint;
    export let mode: "show" | "edit" = "show";

    const dropdownItems = [
        { text: $_("edit"), value: "edit" },
        { text: $_("delete"), value: "delete" },
    ];
</script>

<div
    class="p-4 border border-input-border rounded-md my-2 hover:bg-menu-item-background-hover"
>
    <div class="flex justify-between items-center mb-2">
        <h5>
            <i class="fa fa-{waypoint.icon} mr-2"></i>{waypoint.name}
        </h5>
        {#if mode == "edit"}
            <Dropdown items={dropdownItems} on:change></Dropdown>
        {/if}
    </div>

    {#if waypoint.description}
        <p>{waypoint.description}</p>
    {/if}
    <span class="text-sm text-gray-500">{waypoint.lat}, {waypoint.lon}</span>
</div>
