<script context="module" lang="ts">
    export type SearchItem = {
        text: string;
        description?: string;
        value: any;
        icon: string;
    };
</script>

<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { fade } from "svelte/transition";
    import TextField from "./text_field.svelte";

    export let maxSearchLength: number = 5;
    export let value: string = "";
    export let items: SearchItem[] = [];
    export let placeholder: string = "Search...";
    export let large: boolean = false;
    export let extraClasses: string = "";

    const dispatch = createEventDispatcher();

    let lastSearch: string = "";
    let searching: boolean = false;
    let typingTimer!: any;

    $: dropDownOpen = value.length > 0 && items.length > 0 && searching;

    function onSearchType() {
        clearTimeout(typingTimer);
        if (Math.abs(value.length - lastSearch.length) > maxSearchLength) {
            update(value);
            return;
        }
        typingTimer = setTimeout(() => {
            update(value);
        }, 500);
    }

    function update(q: string) {
        lastSearch = q;
        dispatch("update", q);
    }

    function handleItemClick(item: SearchItem) {
        searching = false;
        dispatch("click", item);
    }

    function clear() {
        value = "";
        update(value);
    }
</script>

<div class="relative {extraClasses}">
    <span class="absolute top-1/2 -translate-y-1/2 left-0 pl-4">
        <i class="fa fa-search" class:text-xl={large}></i>
    </span>
    {#if value.length > 0}
        <button
            class="btn-icon absolute top-1/2 -translate-y-1/2 right-0 mr-2"
            on:click={clear}
            in:fade={{ duration: 150 }}
            out:fade={{ duration: 150 }}
        >
            <i class="fa fa-close text-sm"></i>
        </button>
    {/if}
    <TextField
        type="search"
        name="q"
        autocomplete="off"
        extraClasses={large
            ? "px-12 py-4 text-xl min-w-80 w-[33vw] max-w-[532px] rounded-xl"
            : "px-10"}
        {placeholder}
        bind:value
        on:input={onSearchType}
        on:focusin={() => (searching = true)}
        on:focusout={() => (searching = false)}
    ></TextField>

    {#if dropDownOpen}
        <ul
            class="menu absolute bg-menu-background border border-input-border rounded-xl shadow-md overflow-hidden w-full"
            class:none={!dropDownOpen}
            style="z-index: 1001"
        >
            {#each items as item}
                <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
                <li
                    class="menu-item flex items-center px-4 py-3 cursor-pointer hover:bg-menu-item-background-hover focus:bg-menu-item-background-focus transition-colors"
                    tabindex="0"
                    on:mousedown|stopPropagation={() => handleItemClick(item)}
                    on:keydown|stopPropagation={() => handleItemClick(item)}
                >
                    <slot name="item-header" {item}>
                        <i class="fa fa-{item.icon} mr-6"></i>
                    </slot>

                    <div>
                        <p>{item.text}</p>
                        {#if item.description}
                            <p class="text-sm text-gray-500">
                                {item.description}
                            </p>
                        {/if}
                    </div>
                </li>
            {/each}
        </ul>
    {/if}
</div>
