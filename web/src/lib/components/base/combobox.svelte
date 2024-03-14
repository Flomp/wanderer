<script context="module" lang="ts">
    export type ComboboxItem = {
        text: string;
        value: any;
        icon: string;
    };
</script>

<script lang="ts">
    import { createEventDispatcher, tick } from "svelte";
    import TextField from "./text_field.svelte";

    export let name: string = "";
    export let icon: string = "";
    export let label: string = "";
    export let value: string = "";
    export let items: ComboboxItem[] = [];
    export let placeholder: string = "";
    export let extraClasses: string = "";

    const dispatch = createEventDispatcher();

    let searching: boolean = false;

    $: dropDownOpen = value.length > 0 && items.length > 0 && searching;

    async function onSearchType() {
        await tick();
        const dropdownMenu = document.querySelector(".menu");
        if (dropdownMenu) {
            let dropdownHTML = dropdownMenu.innerHTML;

            for (const li of dropdownMenu.children) {
                const textNode = li.getElementsByTagName("p")[0];
                const text = textNode.innerText.replace(
                    new RegExp(value, "gi"),
                    (match) => `<strong>${match}</strong>`,
                );
                textNode.innerHTML = text;
            }
        }

        update(value);
    }

    function update(q: string) {
        dispatch("update", q);
    }

    function handleItemClick(item: ComboboxItem) {
        value = item.value;
        dispatch("click", item);
    }
</script>

<div class="relative {extraClasses}">
    <TextField
        type="search"
        {name}
        autocomplete="off"
        {icon}
        {label}
        {placeholder}
        bind:value
        on:change
        on:input={onSearchType}
        on:focusin={() => (searching = true)}
        on:focusout={() => (searching = false)}
    ></TextField>

    {#if dropDownOpen}
        <ul
            class="menu absolute bg-menu-background border border-input-border rounded-xl shadow-md w-full max-h-64 overflow-x-hidden overflow-y-scroll"
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
                    <i class="fa fa-{item.icon} mr-6"></i>

                    <p class="text-ellipsis">{item.text}</p>
                </li>
            {/each}
        </ul>
    {/if}
</div>
