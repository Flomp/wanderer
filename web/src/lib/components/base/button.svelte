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
    export let tooltip: string | undefined = undefined;
    export let id: string | undefined = undefined
</script>

<button
    class="relative {extraClasses} flex items-center justify-center"
    class:btn-primary={primary}
    class:btn-secondary={secondary}
    class:btn-large={large}
    class:btn-disabled={disabled || loading}
    disabled={disabled || loading}
    class:tooltip={(tooltip?.length ?? 0) > 0}
    data-title={tooltip}
    on:click
    {type}
    {id}
>
    <div class:invisible={loading}>
        {#if icon}
            <i class="fa fa-{icon} mr-2"></i>
        {/if}
        <slot />
    </div>
    {#if loading}
        <div class="absolute aspect-square spinner"></div>
    {/if}
</button>
