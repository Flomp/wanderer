<script lang="ts">
    import type { List } from "$lib/models/list";
    import Dropdown, { type DropdownItem } from "../base/dropdown.svelte";
    export let list: List;
    export let active: boolean = false;

    const dropdownItems: DropdownItem[] = [
        { text: "Edit", value: "edit" },
        { text: "Delete", value: "delete" },
    ];
</script>

<div
    class="flex items-center gap-6 p-4 hover:bg-menu-item-background-hover rounded-xl transition-colors cursor-pointer"
    class:bg-menu-item-background-hover={active}
>
    <img
        class="w-24 aspect-square rounded-full"
        src={list.avatar}
        alt="avatar"
    />
    <div class="basis-full self-start">
        <div class="flex justify-between items-center">
            <h5 class="text-xl font-semibold">{list.name}</h5>
            <Dropdown items={dropdownItems} on:change></Dropdown>
        </div>
        <p class="text-gray-500 text-sm mr-8">{list.description}</p>
    </div>
</div>
