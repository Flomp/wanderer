<script lang="ts">
    export let value: string | number = "";
    export let placeholder: string = "";
    export let label: string = "";
    export let error: string = "";
    export let icon: string = "";
</script>

<div>
    {#if label.length}
        <p class="text-sm font-medium pb-1">
            {label}
        </p>
    {/if}
    <div class="flex items-center gap-2">
        {#if icon.length > 0}
            <i class="fa fa-{icon}"></i>
        {/if}
        <input
            class="bg-gray-50 border rounded-md p-3 transition-colors focus:border-primary focus:outline-none focus:ring-0 w-full"
            class:border-red-400={error.length > 0}
            class:bg-red-50={error.length > 0}
            type="text"
            bind:value
            on:change
            {placeholder}
        />
    </div>

    <span class="textfield-error text-xs text-red-400">
        {error}
    </span>
</div>
