<script lang="ts">
    import { _ } from "svelte-i18n";

    export let name: string = "";
    export let value: string | number = "";
    export let placeholder: string = "";
    export let rows: number = 3;
    export let label: string = "";
    export let error: string | string[] | null = "";
    export let extraClasses: string = "";
</script>

<div>
    {#if label.length}
        <p class="text-sm font-medium mb-1">
            {label}
        </p>
    {/if}
    <textarea
        {name}
        class="bg-input-background border border-input-border rounded-md p-3 resize-none transition-colors focus:border-input-border-focus focus:outline-none focus:ring-0 w-full {extraClasses}"
        {rows}
        {placeholder}
        class:border-red-400={(error?.length ?? 0) > 0}
        class:bg-input-background-error={(error?.length ?? 0) > 0}
        bind:value
        on:change
    />
    {#if error}
        <span class="textfield-error text-xs text-red-400">
            {error instanceof Array ? $_(error[0]) : error}
        </span>
    {/if}
</div>
