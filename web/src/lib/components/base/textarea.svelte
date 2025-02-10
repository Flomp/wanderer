<script lang="ts">
    import { createBubbler } from 'svelte/legacy';

    const bubble = createBubbler();
    import { _ } from "svelte-i18n";

    interface Props {
        name?: string;
        value?: string | number;
        placeholder?: string;
        rows?: number;
        label?: string;
        error?: string | string[] | null;
        extraClasses?: string;
    }

    let {
        name = "",
        value = $bindable(""),
        placeholder = "",
        rows = 3,
        label = "",
        error = "",
        extraClasses = ""
    }: Props = $props();
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
        onchange={bubble('change')}
></textarea>
    {#if error}
        <span class="textfield-error text-xs text-red-400">
            {error instanceof Array ? $_(error[0]) : error}
        </span>
    {/if}
</div>
