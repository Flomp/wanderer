<script lang="ts">
    export let name: string = "";
    export let value: any;
    export let items: { text: string; value: any }[] = [];
    export let label: string = "";
</script>

<div>
    {#if label.length}
        <label for={name} class="text-sm font-medium pb-1">
            {label}
        </label>
    {/if}
    <select
        {name}
        class="bg-gray-50 h-10 w-full px-4 border-r-8 border-transparent outline outline-1 outline-gray-200 rounded-md focus:outline-primary transition-colors"
        bind:value
    >
        {#each items as item}
            <option value={item.value}>{item.text}</option>
        {/each}
    </select>
</div>
