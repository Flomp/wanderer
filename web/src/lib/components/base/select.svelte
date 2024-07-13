<script context="module" lang="ts">
    export type SelectItem = {
        text: string;
        value: any;
    };
</script>

<script lang="ts">
    import { createEventDispatcher } from "svelte";

    export let name: string = "";
    export let value: any;
    export let items: SelectItem[] = [];
    export let label: string = "";
    export let disabled: boolean = false;

    const dispatch = createEventDispatcher();

    function onChange(target: any) {
        dispatch("change", target?.value);
    }
</script>

<div>
    {#if label.length}
        <label for={name} class="text-sm font-medium pb-1">
            {label}
        </label>
    {/if}
    <select
        {name}
        class="block bg-input-background h-10 px-4 border-r-8 border-transparent outline outline-1 outline-input-border rounded-md focus:outline-input-border-focus transition-colors"
        class:text-gray-500={disabled}
        {disabled}
        bind:value
        on:change={(e) => onChange(e.target)}
    >
        {#each items as item}
            <option value={item.value}>{item.text}</option>
        {/each}
    </select>
</div>
