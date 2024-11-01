<script lang="ts">
    export let label: string = "";
    export let name: string = "";
    export let accept: string = "*";

    export let files: FileList | null = null;
</script>

<div>
    {#if label.length}
        <label for={name} class="text-sm font-medium pb-1">
            {label}
        </label>
    {/if}
    <input
        {name}
        {accept}
        bind:files
        class="cursor-pointer bg-input-background border border-input-border rounded-md p-3 transition-colors focus:border-input-border-focus focus:outline-none focus:ring-0 w-full file:rounded-lg file:px-4 file:py-2 file:border file:border-input-border-focus file:font-medium file:mr-4"
        type="file"
        on:change
    />
</div>
