<script lang="ts">
    import { createEventDispatcher } from "svelte";

    export let items: { text: string; value: any }[] = [];

    const dispatch = createEventDispatcher();

    let isOpen = false;

    function toggleMenu(e: MouseEvent) {
        e.stopPropagation();
        e.preventDefault();
        isOpen = !isOpen;
    }

    function closeMenu() {
        isOpen = false;
    }

    function handleItemClick(item: { text: string; value: any }) {
        dispatch("change", item);
        closeMenu();
    }
</script>

<svelte:window on:mouseup={() => (isOpen = false)} />

<div class="dropdown relative">
    <button
        class="hover:bg-gray-100 w-8 h-8 rounded-full"
        on:click={toggleMenu}
        type="button"
    >
        <i class="fa fa-ellipsis-vertical"></i>
    </button>

    {#if isOpen}
        <ul
            class="menu absolute bg-white border rounded-l-xl rounded-b-xl shadow-md z-10 right-0 overflow-hidden"
            class:none={isOpen}
        >
            {#each items as item}
                <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
                <li
                    class="menu-item p-4 cursor-pointer hover:bg-gray-100 focus:bg-gray-200 transition-colors"
                    tabindex="0"
                    on:mouseup|stopPropagation={() => handleItemClick(item)}
                    on:keydown|stopPropagation={() => handleItemClick(item)}
                >
                    {item.text}
                </li>
            {/each}
        </ul>
    {/if}
</div>

<style>
</style>
