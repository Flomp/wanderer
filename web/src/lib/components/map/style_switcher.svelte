<script lang="ts">
    import type { StyleSwitcherControlOptions } from "$lib/vendor/maplibre-style-switcher/style-switcher-control";
    import RadioGroup from "../base/radio_group.svelte";
    import { _ } from "svelte-i18n";

    interface Props {
        settings: StyleSwitcherControlOptions;
    }

    let { settings }: Props = $props();

    let container: HTMLDivElement;

    const mapStyles = $derived(settings.styles);

    let showSwitcher: boolean = $state(false);

    function toggleSwitcher() {
        showSwitcher = !showSwitcher;
    }

    function hideSwitcher() {
        showSwitcher = false;
    }

    export function getElement() {
        return container;
    }
</script>

<div
    bind:this={container}
    class="maplibregl-ctrl maplibregl-ctrl-group relative z-10"
>
    <button onclick={toggleSwitcher} aria-label="toggle style switcher">
        <i class="fa fa-layer-group text-black"></i>
    </button>

    <div
        class="absolute bg-menu-background rounded-lg px-4 py-2 mt-2 shadow-xl space-y-3"
        class:hidden={!showSwitcher}
        style="transform: translateX(calc(-100% + 29px))"
    >
        <div class="flex items-center gap-x-12">
            <span class="font-semibold whitespace-nowrap text-base"
                >Map style</span
            >
            <button
                class="fa fa-close !rounded-full"
                onclick={hideSwitcher}
                aria-label="close style switcher"
            ></button>
        </div>
        <RadioGroup
            name="completed"
            items={mapStyles}
            bind:selected={settings.state.selectedStyle}
            onchange={(style) => settings.onMapStyleSwitch(settings.styles.indexOf(style), style)}
        ></RadioGroup>
        <hr class="border-input-border" />
        <span class="font-semibold whitespace-nowrap text-base">Layers</span>
        {#each settings.layers as l, i}
            <div class="flex items-center mt-2 mb-4">
                <input
                    id="{l.text}-layer-checkbox"
                    type="checkbox"
                    checked={settings.state.selectedLayers[l.text]}
                    class="w-4 h-4 bg-input-background accent-primary border-input-border focus:ring-input-ring focus:ring-2"
                    onchange={(e) => {
                        const checked = (e.target as HTMLInputElement).checked;
                        settings.onLayerChange(checked, l);
                    }}
                />

                <label for="{l.text}-layer-checkbox" class="ms-2 text-sm"
                    >{$_(l.text)}</label
                >
            </div>
        {/each}
    </div>
</div>
