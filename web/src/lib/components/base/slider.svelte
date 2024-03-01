<script lang="ts">
    import noUiSlider from "nouislider";
    import "nouislider/dist/nouislider.css";
    import { createEventDispatcher, onMount } from "svelte";

    export let minValue = 0;
    export let maxValue = 100;

    export let currentValue = maxValue / 2;

    let sliderContainer: any;

    const dispatch = createEventDispatcher();

    onMount(() => {
        const updateValues = (values: string[]) => {
            currentValue = parseFloat(values[0]);
        };

        noUiSlider.create(sliderContainer, {
            start: currentValue,
            connect: [true, false],
            range: {
                min: minValue,
                max: maxValue,
            },
        });

        sliderContainer.noUiSlider.on("update", updateValues);

        sliderContainer.noUiSlider.on("set", () => {
            dispatch("set", currentValue);
        });
    });
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
