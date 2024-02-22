<script lang="ts">
    export let name: string = "";
    export let value: string | number = "";
    export let placeholder: string = "";
    export let rows: number = 3;
    export let label: string = "";
    export let error: string = "";
</script>

<div>
    {#if label.length}
        <p class="text-sm font-medium mb-1">
            {label}
        </p>
    {/if}
    <textarea
        {name}
        class="bg-input-background border border-input-border rounded-md p-3 resize-none transition-colors focus:border-input-border-focus focus:outline-none focus:ring-0 w-full"
        {rows}
        {placeholder}
        class:border-red-400={error.length > 0}
        class:bg-red-50={error.length > 0}
        bind:value
        on:change
    />
    <span class="textfield-error text-xs text-red-400">
        {error}
    </span>
</div>
