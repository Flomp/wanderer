<script lang="ts">
    import { goto } from "$app/navigation";
    import type { DropdownItem } from "../base/dropdown.svelte";
    import { _ } from "svelte-i18n";
    export let selected: number;

    const items: DropdownItem[] = [
        { text: $_("my-account"), value: "/settings/account" },
        { text: $_("display"), value: "/settings/display" },
        { text: `${$_("import")}/${$_("export")}`, value: "/settings/export" },
        { text: $_('help'), value: "https://wanderer.to" },
    ];

    function handleItemClick(item: DropdownItem) {
        if (item.text != "Help") {
            goto(item.value);
        } else {
            window.open(item.value, "_blank");
        }
    }
</script>

<div
    class="grid grid-cols-1 md:grid-cols-[256px_1fr] max-w-4xl mx-4 md:mx-auto gap-x-12 items-start"
>
    <menu class="p-4 border border-input-border rounded-xl mb-4">
        {#each items as item, i}
            <li
                class="menu-item flex items-center px-4 py-3 my-1 cursor-pointer hover:bg-menu-item-background-hover focus:bg-menu-item-background-focus transition-colors rounded-md"
                class:bg-menu-item-background-hover={i == selected}
                role="presentation"
                on:mousedown|stopPropagation={() => handleItemClick(item)}
            >
                {item.text}
            </li>
        {/each}
    </menu>
    <div class="settings-content">
        <slot></slot>
    </div>
</div>
