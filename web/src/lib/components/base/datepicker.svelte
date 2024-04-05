<script lang="ts">
    export let name: string = "";
    export let value: string | Date = "";
    export let label: string = "";
    export let error: string = "";
</script>

<div>
    {#if label.length}
        <p class="text-sm font-medium pb-1">
            {label}
        </p>
    {/if}
    <div class="flex items-center gap-2">
        <input
            {name}
            class="bg-input-background border border-input-border rounded-md p-3 transition-colors focus:border-input-border-focus focus:outline-none focus:ring-0 w-full"
            class:border-red-400={error.length > 0}
            class:bg-red-50={error.length > 0}
            type="date"
            bind:value
            on:change
        />
    </div>

    <span class="textfield-error text-xs text-red-400">
        {error}
    </span>
</div>
