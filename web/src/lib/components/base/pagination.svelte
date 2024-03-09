<script lang="ts">
    import { range } from "$lib/util/array_util";
    import { createEventDispatcher } from "svelte";

    export let page: number;
    export let totalPages: number;

    const maxPagesToRender = 6;
    const dispatch = createEventDispatcher();

    function update(clickedPage: number) {
        if (page !== clickedPage) {
            dispatch("pagination", clickedPage);
        }
    }

    $: numberGenerator = () => {
        if (totalPages < 4) {
            return [];
        }
        if (totalPages <= maxPagesToRender) {
            return range(totalPages - 2, 2);
        }
        if (page < maxPagesToRender - 2) {
            return range(maxPagesToRender - 1, 2);
        }
        if (page > totalPages - 3) {
            return range(totalPages - 2, totalPages - 5);
        }
        if (page > 2 && page <= totalPages - 3) {
            return range(page + Math.max(1, maxPagesToRender - 5), page - 2);
        }
        return [];
    };
</script>

<div class="flex gap-x-4 justify-center">
    {#if totalPages > 1}
        <button
            class:text-gray-500={page == 1}
            disabled={page == 1}
            on:click={() => update(page - 1)}
            ><i class="fa fa-caret-left"></i></button
        >
        <button
            class="w-8 aspect-square rounded-lg transition-colors border-input-border hover:bg-menu-item-background-hover"
            class:border={page == 1}
            on:click={() => update(1)}>1</button
        >
        {#if totalPages > maxPagesToRender && page >= maxPagesToRender - 2 && page > 2}
            <span>...</span>
        {:else if totalPages > 2}
            <button
                class="w-8 aspect-square rounded-lg transition-colors border-input-border hover:bg-menu-item-background-hover"
                class:border={page == 2}
                on:click={() => update(2)}>2</button
            >
        {/if}
        {#each numberGenerator() as i}
            <button
                class="w-8 aspect-square rounded-lg transition-colors border-input-border hover:bg-menu-item-background-hover"
                class:border={page == i + 1}
                on:click={() => update(i + 1)}>{i + 1}</button
            >
        {/each}
        {#if totalPages > maxPagesToRender && page < totalPages - 3}
            <span>...</span>
        {:else if totalPages > 3}
            <button
                class="w-8 aspect-square rounded-lg transition-colors border-input-border hover:bg-menu-item-background-hover"
                class:border={page == totalPages - 1}
                on:click={() => update(totalPages - 1)}>{totalPages - 1}</button
            >
        {/if}
        <button
            class="w-8 aspect-square rounded-lg transition-colors border-input-border hover:bg-menu-item-background-hover"
            class:border={page == totalPages}
            on:click={() => update(totalPages)}>{totalPages}</button
        >
        <button
            class:text-gray-500={page == totalPages}
            disabled={page == totalPages}
            on:click={() => update(page + 1)}
            ><i class="fa fa-caret-right"></i></button
        >
    {/if}
</div>
