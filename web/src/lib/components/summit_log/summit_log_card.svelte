<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import { _ } from "svelte-i18n";
    import Dropdown, { type DropdownItem } from "../base/dropdown.svelte";

    export let log: SummitLog;
    export let mode: "show" | "edit" = "show";

    const dropdownItems: DropdownItem[] = [
        { text: $_("edit"), value: "edit" },
        { text: $_("delete"), value: "delete" },
    ];
</script>

<div class="p-4 my-2 border border-input-border rounded-xl">
    <div class="flex justify-between items-center mb-2">
        <h5 class="font-medium mr-2">
            {new Date(log.date).toLocaleDateString(undefined, {
                month: "2-digit",
                day: "2-digit",
                year: "numeric",
                timeZone: "UTC",
            })}
        </h5>

        {#if mode == "edit"}
            <Dropdown items={dropdownItems} on:change></Dropdown>
        {/if}
    </div>
    <span>{log.text}</span>
</div>
