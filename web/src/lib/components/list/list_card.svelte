<script lang="ts">
    import type { List } from "$lib/models/list";
    import { getFileURL } from "$lib/util/file_util";
    import Dropdown, { type DropdownItem } from "../base/dropdown.svelte";
    import { _ } from "svelte-i18n";
    export let list: List;
    export let active: boolean = false;

    const dropdownItems: DropdownItem[] = [
        { text: $_("edit"), value: "edit" },
        { text: $_("delete"), value: "delete" },
    ];
</script>

<div
    class="flex items-center gap-6 p-4 hover:bg-menu-item-background-hover rounded-xl transition-colors cursor-pointer"
    class:bg-menu-item-background-hover={active}
>
    {#if list.avatar}
        <img
            class="w-16 md:w-24 aspect-square rounded-full object-cover"
            src={getFileURL(list, list.avatar)}
            alt="avatar"
        />
    {:else}
        <div
            class="flex w-16 md:w-24 aspect-square shrink-0 items-center justify-center"
        >
            <i class="fa fa-table-list text-5xl"></i>
        </div>
    {/if}
    <div class="self-start min-w-0 w-full">
        <div class="flex justify-between items-center ">
            <h5 class="text-xl font-semibold overflow-hidden overflow-ellipsis">{list.name}</h5>
            <Dropdown items={dropdownItems} on:change></Dropdown>
        </div>
        <p class="text-gray-500 text-sm mr-8">{list.description}</p>
    </div>
</div>
