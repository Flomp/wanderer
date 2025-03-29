<script lang="ts">
    import noUiSlider from "nouislider";
    import "nouislider/dist/nouislider.css";
    import { onMount } from "svelte";
    import { type Options as SliderOptions } from "nouislider";

    interface Props {
        minValue?: number;
        maxValue?: number;
        currentValue?: any;
        onset?: (value: number) => void;
    }

    let {
        minValue = 0,
        maxValue = 100,
        currentValue = $bindable(maxValue / 2),
        onset,
        ...sliderOptions
    }: Props & Partial<SliderOptions> = $props();

    let sliderContainer: any = $state();

    onMount(() => {
        const updateValues = (values: string[]) => {
            currentValue = parseFloat(values[0]);
        };

        noUiSlider.create(sliderContainer, {
            start: currentValue,
            connect: [true, false],
            ...sliderOptions,
            range: {
                min: minValue,
                max: maxValue,
            },
        });

        sliderContainer.noUiSlider.on("update", updateValues);

        sliderContainer.noUiSlider.on("set", () => {
            onset?.(currentValue);
        });
    });

    export function set(value: number) {
        sliderContainer.noUiSlider.set(value)
    }
</script>

<div class="my-4" id="slider" bind:this={sliderContainer}></div>

<style>
    :global(.noUi-target) {
        @apply border border-input-border shadow-none;
    }
    :global(.noUi-horizontal) {
        height: 6px;
    }
    :global(.noUi-connect) {
        @apply bg-primary;
    }

    :global(.noUi-horizontal .noUi-handle) {
        @apply rounded-full w-7 h-7 -top-3;
    }

    :global(.noUi-handle::before) {
        @apply content-none;
    }

    :global(.noUi-handle::after) {
        @apply content-none;
    }
</style>
