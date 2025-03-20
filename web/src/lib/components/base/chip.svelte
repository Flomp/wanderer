<script lang="ts">
    import type { MouseEventHandler } from "svelte/elements";

    interface Props {
        text: string;
        primary?: boolean;
        closable?: boolean;
        onclick?: MouseEventHandler<HTMLButtonElement>;
    }

    let { text, primary = true, closable = false, onclick }: Props = $props();
</script>

<div
    class="{primary
        ? 'bg-primary text-white'
        : 'border border-input-border bg-menu-item-background-hover'} px-2 py-1 rounded-full flex items-center gap-1"
>
    <span class="text-sm">{text}</span>
    {#if closable}
        <button
            type="button"
            aria-label="Close"
            {onclick}
            class="text-white hover:bg-primary-hover rounded-full w-4 h-4 flex items-center justify-center"
        >
            <i class="fa fa-close text-xs"></i>
        </button>
    {/if}
</div>
