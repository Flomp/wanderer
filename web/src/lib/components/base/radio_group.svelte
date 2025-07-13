<script module lang="ts">
    export type RadioItem = {
        text: string;
        value: any;
        icon?: string;
        description?: string;
    };
</script>

<script lang="ts">
    interface Props {
        items: RadioItem[];
        name: string;
        selected?: number;
        onchange?: (item: RadioItem) => void
    }

    let { items, name, selected = $bindable(0), onchange }: Props = $props();


    function handleRadioChange(radioIndex: number) {
        selected = radioIndex;
        onchange?.(items[selected]);
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
