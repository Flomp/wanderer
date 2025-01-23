<script module lang="ts">
    export type ComboboxItem = {
        text: string;
        value: any;
        icon: string;
    };
</script>

<script lang="ts">
    import { createEventDispatcher, tick } from "svelte";
    import type { ChangeEventHandler } from "svelte/elements";
    import TextField from "./text_field.svelte";

    interface Props {
        name?: string;
        icon?: string;
        label?: string;
        value?: string;
        items?: ComboboxItem[];
        placeholder?: string;
        extraClasses?: string;
        onchange?: ChangeEventHandler<HTMLInputElement>;
    }

    let {
        name = "",
        icon = "",
        label = "",
        value = $bindable(""),
        items = [],
        placeholder = "",
        extraClasses = "",
        onchange,
    }: Props = $props();

    const dispatch = createEventDispatcher();

    let searching: boolean = $state(false);

    let dropDownOpen = $derived(
        value.length > 0 && items.length > 0 && searching,
    );

    async function onSearchType() {
        await tick();
        const dropdownMenu = document.querySelector(".menu");
        if (dropdownMenu) {
            for (let i = 0; i < dropdownMenu.children.length; i++) {
                const li = dropdownMenu.children[i];
                const textNode = li.getElementsByTagName("p")[0];
                textNode.innerHTML = items[i].text;

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

    function handleItemClick(e: Event, item: ComboboxItem) {
        e.stopPropagation();
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
        {onchange}
        oninput={onSearchType}
        onfocusin={() => (searching = true)}
        onfocusout={() => (searching = false)}
    ></TextField>

    {#if dropDownOpen}
        <ul
            class="menu absolute bg-menu-background border border-input-border rounded-xl shadow-md w-full max-h-64 overflow-x-hidden overflow-y-scroll"
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
                    <i class="fa fa-{item.icon} mr-6"></i>

                    <p class="text-ellipsis">{item.text}</p>
                </li>
            {/each}
        </ul>
    {/if}
</div>
