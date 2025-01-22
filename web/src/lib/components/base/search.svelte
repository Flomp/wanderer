<!-- @migration-task Error while migrating Svelte code: This migration would change the name of a slot making the component unusable -->
<script module lang="ts">
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

    interface Props {
        maxSearchLength?: number;
        value?: string;
        items?: SearchItem[];
        placeholder?: string;
        large?: boolean;
        extraClasses?: string;
        label?: string;
        clearAfterSelect?: boolean;
        prepend?: import('svelte').Snippet<[any]>;
    }

    let {
        maxSearchLength = 5,
        value = $bindable(""),
        items = $bindable([]),
        placeholder = "Search...",
        large = false,
        extraClasses = "",
        label = "",
        clearAfterSelect = true,
        prepend
    }: Props = $props();

    const dispatch = createEventDispatcher();

    let lastSearch: string = "";
    let searching: boolean = $state(false);
    let typingTimer!: any;

    let dropDownOpen = $derived(value.length > 0 && items.length > 0 && searching);

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

    function handleItemClick(e: Event, item: SearchItem) {
        e.stopPropagation();
        searching = false;
        dispatch("click", item);
        if (clearAfterSelect) {
            clear();
        }
    }

    function clear() {
        value = "";
        items = [];
        update(value);
    }
</script>

<div class="relative {extraClasses}">
    <span
        class="absolute {large
            ? 'bottom-[31px]'
            : 'bottom-[25px]'} translate-y-1/2 left-0 pl-4"
    >
        <i class="fa fa-search" class:text-xl={large}></i>
    </span>
    {#if value.length > 0}
        <button
            aria-label="Clear"
            type="button"
            class="btn-icon absolute {large
                ? 'bottom-[31px]'
                : 'bottom-[25px]'} translate-y-1/2 right-0 mr-2"
            onclick={clear}
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
        {label}
        extraClasses={large
            ? "px-12 py-4 text-xl min-w-80 w-[33vw] max-w-[532px] rounded-xl"
            : "px-10"}
        {placeholder}
        bind:value
        oninput={onSearchType}
        onfocusin={() => (searching = true)}
        onfocusout={() => (searching = false)}
    ></TextField>

    {#if dropDownOpen}
        <ul
            class="menu absolute bg-menu-background border border-input-border rounded-xl shadow-md overflow-hidden w-full"
            class:none={!dropDownOpen}
            style="z-index: 1001"
        >
            {#each items as item}
                <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
                <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
                <li
                    class="menu-item flex items-center px-4 py-3 cursor-pointer hover:bg-menu-item-background-hover focus:bg-menu-item-background-focus transition-colors"
                    tabindex="0"
                    onmousedown={(e) => handleItemClick(e, item)}
                    onkeydown={(e) => handleItemClick(e, item)}
                >
                    {#if prepend}{@render prepend({ item, })}{:else}
                        <i class="fa fa-{item.icon} mr-6"></i>
                    {/if}

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
