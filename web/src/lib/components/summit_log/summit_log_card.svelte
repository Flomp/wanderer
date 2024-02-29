<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import { format, parse } from "date-fns";
    import Dropdown, { type DropdownItem } from "../base/dropdown.svelte";
    import { _ } from "svelte-i18n";

    export let log: SummitLog;
    export let mode: "show" | "edit" = "show";

    const dropdownItems: DropdownItem[] = [
        { text: $_("edit"), value: "edit" },
        { text: $_("delete"), value: "delete" },
    ];
</script>

<div class="p-4 my-2 border border-input-border rounded-xl">
    <div class="flex justify-between items-center mb-2">
        <h5 class="font-medium mr-2">{format(log.date, "dd.MM.yyyy")}</h5>

        {#if mode == "edit"}
            <Dropdown items={dropdownItems} on:change></Dropdown>
        {/if}
    </div>
    <span>{log.text}</span>
</div>
