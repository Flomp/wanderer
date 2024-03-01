<script lang="ts">
    import * as noUiSlider from "nouislider";
    import "nouislider/dist/nouislider.css";
    import { createEventDispatcher, onMount } from "svelte";

    export let minValue = 0;
    export let maxValue = 100;

    export let currentMin = minValue;
    export let currentMax = maxValue;

    let sliderContainer: any;

    const dispatch = createEventDispatcher();

    onMount(() => {
        const updateValues = (values: string[]) => {
            currentMin = parseFloat(values[0]);
            currentMax = parseFloat(values[1]);
        };

        noUiSlider.create(sliderContainer, {
            start: [currentMin, currentMax],
            connect: true,
            range: {
                min: minValue,
                max: maxValue,
            },
        });

        sliderContainer.noUiSlider.on("update", updateValues);

        sliderContainer.noUiSlider.on("set", () => {
            dispatch("set", [currentMin, currentMax]);
        });
    });
</script>

<div class="my-4" id="slider" bind:this={sliderContainer}></div>

<style>
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
