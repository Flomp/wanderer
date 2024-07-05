<script lang="ts">
    export let name: string = "";
    export let value: string | number = "";
    export let placeholder: string = "";
    export let disabled: boolean = false;
    export let label: string = "";
    export let error: string = "";
    export let icon: string = "";
    export let extraClasses: string = "";
    export let type: "text" | "password" | "search" = "text";
    export let autocomplete: "on" | "off" = "on";
    function typeAction(node: HTMLInputElement) {
        node.type = type;
    }
</script>

<div>
    {#if label.length}
        <label for={name} class="text-sm font-medium pb-1">
            {label}
        </label>
    {/if}
    <div class="flex items-center gap-2">
        {#if icon.length > 0}
            <i class="fa fa-{icon}"></i>
        {/if}
        <input
            {name}
            class="bg-input-background border border-input-border rounded-md p-3 transition-colors focus:border-input-border-focus focus:outline-none focus:ring-0 w-full {extraClasses}"
            class:border-red-400={error.length > 0}
            class:bg-input-background-error={error.length > 0}
            class:text-gray-500={disabled}
            {disabled}
            {autocomplete}
            use:typeAction
            bind:value
            on:change
            on:input
            on:focusin
            on:focusout
            {placeholder}
        />
    </div>

    <span class="textfield-error text-xs text-red-400">
        {error}
    </span>
</div>
