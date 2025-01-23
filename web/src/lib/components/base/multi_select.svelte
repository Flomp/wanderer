<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";
    import type { SelectItem } from "./select.svelte";

    interface Props {
        items?: SelectItem[];
        value?: SelectItem[];
        label?: string;
        name?: string;
        placeholder?: string;
    }

    let {
        items = [],
        value = $bindable([]),
        label = "",
        name = "",
        placeholder = "",
    }: Props = $props();

    let showDropdown = $state(false);
    let dropdownRef = $state();

    const dispatch = createEventDispatcher();

    function toggleItem(item: SelectItem) {
        if (value.includes(item)) {
            value = value.filter((i) => i !== item);
        } else {
            value = [...value, item];
        }

        dispatch("change", value);
    }

    function removeItem(e: Event, item: SelectItem) {
        e.stopPropagation();
        value = value.filter((i) => i !== item);
        dispatch("change", value);
    }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<div class="relative max-w-full">
    {#if label.length}
        <label for={name} class="text-sm font-medium pb-1 block">
            {label}
        </label>
    {/if}
    <div
        role="presentation"
        class="min-w-44 w-full flex flex-wrap items-center gap-2 border border-input-border bg-input-background min-h-[50px] p-3 rounded-md transition-colors focus:border-input-border-focus focus:outline-none focus:ring-0"
        onclick={() => (showDropdown = !showDropdown)}
    >
        {#if value.length === 0}
            <span class="text-gray-400">{placeholder}</span>
        {/if}
        {#each value as item}
            <div
                class="bg-primary text-white px-2 py-1 rounded-full flex items-center gap-1"
            >
                <span class="text-sm">{$_(item.text)}</span>
                <button
                    aria-label="Close"
                    onclick={(e) => removeItem(e, item)}
                    class="text-white hover:bg-primary-hover rounded-full w-4 h-4 flex items-center justify-center"
                >
                    <i class="fa fa-close"></i>
                </button>
            </div>
        {/each}
    </div>

    <!-- Dropdown menu -->
    {#if showDropdown}
        <div
            bind:this={dropdownRef}
            class="absolute z-10 mt-1 w-full bg-menu-background border border-input-border rounded-md max-h-40 overflow-y-auto"
        >
            {#each items as item}
                <button
                    onclick={() => toggleItem(item)}
                    class="px-3 py-2 hover:bg-menu-item-background-hover cursor-pointer flex justify-between items-center w-full"
                >
                    <span>{$_(item.text)}</span>
                    {#if value.includes(item)}
                        <div class="ml-auto">
                            <i class="fa fa-check"></i>
                        </div>
                    {/if}
                </button>
            {/each}
        </div>
    {/if}
</div>
