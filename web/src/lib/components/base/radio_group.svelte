<script module lang="ts">
    export type RadioItem = {
        text: string;
        value: string;
        icon?: string;
        description?: string;
    };
</script>

<script lang="ts">
    import { createEventDispatcher } from "svelte";

    interface Props {
        items: RadioItem[];
        name: string;
        selected?: number;
    }

    let { items, name, selected = $bindable(0) }: Props = $props();

    const dispatch = createEventDispatcher();

    function handleRadioChange(radioIndex: number) {
        selected = radioIndex;
        dispatch("change", items[selected]);
    }
</script>

{#each items as item, i}
    <div class="flex items-center mb-4">
        <input
            id="{name}-radio-{i}"
            name="{name}-radio"
            type="radio"
            checked={i == selected}
            value={item.value}
            class="w-4 h-4 accent-primary border-input-border focus:ring-input-ring focus:ring-2"
            onchange={() => handleRadioChange(i)}
        />
        <div class="ms-2 text-sm">
            <label for="{name}-radio-{i}"
                >{item.text}
                {#if item.description}
                    <p class="text-gray-500">{item.description}</p>
                {/if}
            </label>
        </div>
    </div>
{/each}
