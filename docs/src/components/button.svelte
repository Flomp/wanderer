<script lang="ts">
    export let type: "button" | "submit" | "reset" | null | undefined =
        undefined;
    export let icon: string = "";
    export let extraClasses: string = "";
    export let primary: boolean = false;
    export let secondary: boolean = false;
    export let large: boolean = false;
    export let loading: boolean = false;
    export let disabled: boolean = false;
    export let tooltip: string = "";
</script>

<button
    class="{extraClasses} flex items-center justify-center"
    class:btn-primary={primary}
    class:btn-secondary={secondary}
    class:btn-large={large}
    class:btn-disabled={disabled || loading}
    disabled={disabled || loading}
    class:tooltip={tooltip.length > 0}
    data-title={tooltip}
    on:click
    {type}
>
    {#if !loading}
        {#if icon}
            <i class="fa fa-{icon} mr-2"></i>
        {/if}
        <slot />
    {:else}
        <div class="spinner"></div>
    {/if}
</button>
