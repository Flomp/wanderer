<script lang="ts">
    import { createEventDispatcher } from "svelte";

    export let items: { text: string; value: any; icon?: string }[] = [];
    export let size: string = "regular";

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
    <button class="flex items-center justify-center" on:click={toggleMenu} type="button">
        <slot>
            <i
                class="fa fa-ellipsis-vertical text-{size} hover:bg-gray-300 hover:bg-opacity-50 px-[14px] py-2 rounded-full "
            ></i>
        </slot>
    </button>
    {#if isOpen}
        <ul
            class="menu absolute bg-white border rounded-l-xl rounded-b-xl shadow-md right-0 overflow-hidden text-black"
            class:none={isOpen}
            style="z-index: 1001"
        >
            {#each items as item}
                <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
                <li
                    class="menu-item flex items-center p-4 cursor-pointer hover:bg-gray-100 focus:bg-gray-200 transition-colors"
                    tabindex="0"
                    on:mouseup|stopPropagation={() => handleItemClick(item)}
                    on:keydown|stopPropagation={() => handleItemClick(item)}
                >
                    {#if item.icon}
                        <i class="fa fa-{item.icon} mr-6"></i>
                    {/if}
                    {item.text}
                </li>
            {/each}
        </ul>
    {/if}
</div>

<style>
</style>
