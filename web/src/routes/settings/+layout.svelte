<script lang="ts">
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import type { SelectItem } from "$lib/components/base/select.svelte";
    import { _ } from "svelte-i18n";

    const settingsLinks: SelectItem[] = [
        { text: $_("profile"), value: "/settings/profile" },
        { text: $_("my-account"), value: "/settings/account" },
        {
            text: $_("language") + " & " + $_("units"),
            value: "/settings/language",
        },
        { text: $_("map"), value: "/settings/map" },
        { text: `${$_("import")}/${$_("export")}`, value: "/settings/export" },
        {
            text: $_("help"),
            value: "https://wanderer.to/getting-started/installation",
        },
    ];

    $: activeIndex = settingsLinks.findIndex(
        (l) => l.value === $page.url.pathname,
    );

    function handleItemClick(item: SelectItem) {
        if (item.value != settingsLinks.at(-1)?.value) {
            goto(item.value);
        } else {
            window.open(item.value, "_blank");
        }
    }
</script>

<div
    class="grid grid-cols-1 md:grid-cols-[256px_1fr] max-w-6xl mx-4 md:mx-auto gap-x-12 items-start"
>
    <menu class="p-4 border border-input-border rounded-xl mb-4">
        {#each settingsLinks as item, i}
            <li
                class="menu-item flex items-center px-4 py-3 my-1 cursor-pointer hover:bg-menu-item-background-hover focus:bg-menu-item-background-focus transition-colors rounded-md"
                class:bg-menu-item-background-hover={i == activeIndex}
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
