<script lang="ts">
    import type {
        RoutingOptions,
        ValhallaBicycleCostingOptions
    } from "$lib/models/valhalla";
    import { formatSpeed } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import { slide } from "svelte/transition";
    import Select, { type SelectItem } from "../base/select.svelte";
    import Slider from "../base/slider.svelte";
    import Toggle from "../base/toggle.svelte";
    interface Props {
        options: RoutingOptions;
    }

    let { options = $bindable() }: Props = $props();

    const modesOfTransport: SelectItem[] = [
        { text: $_("hiking"), value: "pedestrian" },
        { text: $_("cycling"), value: "bicycle" },
        { text: $_("driving"), value: "auto" },
    ];

    const bikeTypes: SelectItem[] = [
        { text: $_("hybrid"), value: "Hybrid" },
        { text: $_("road"), value: "Road" },
        { text: $_("cross"), value: "Cross" },
        { text: $_("mountain"), value: "Mountain" },
    ];

    if (!options.pedestrianOptions) {
        options.pedestrianOptions = {
            max_hiking_difficulty: 6,
            walking_speed: 5.1,
            use_hills: 1,
            shortest: false,
        };
    }

    if (!options.bicycleOptions) {
        options.bicycleOptions = {
            bicycle_type: "Hybrid",
            cycling_speed: 20,
            use_roads: 0.5,
            use_hills: 0.5,
            avoid_bad_surfaces: 0.25,
            shortest: false,
        };
    }

    if (!options.autoOptions) {
        options.autoOptions = {
            width: 1.6,
            height: 1.9,
            top_speed: 140,
            fixed_speed: 0,
            shortest: false,
        };
    }

    $effect(() => {
        if(!options.autoRouting) {
            showSettings = false;
        }
    })

    let showSettings = $state(false);

    // svelte-ignore non_reactive_update
    let cycleSpeedSlider: Slider;

    function adjustSpeeddependingOnBikeType(
        type: ValhallaBicycleCostingOptions["bicycle_type"],
    ) {
        switch (type) {
            case "City":
            case "Hybrid":
                cycleSpeedSlider?.set(18);
                break;
            case "Road":
                cycleSpeedSlider?.set(25);
                break;
            case "Cross":
                cycleSpeedSlider?.set(20);
                break;
            case "Mountain":
                cycleSpeedSlider?.set(16);
                break;
            default:
                break;
        }
    }
</script>

<div class="p-4 my-2 rounded-xl bg-background space-y-4">
    <Toggle bind:value={options.autoRouting} label={$_("enable-auto-routing")}
    ></Toggle>
    <div class="flex items-center gap-4">
        <Select
            items={modesOfTransport}
            bind:value={options.modeOfTransport}
            disabled={!options.autoRouting}
        ></Select>
        <button
            class="btn-icon"
            type="button"
            disabled={!options.autoRouting}
            onclick={() => (showSettings = !showSettings)}
            aria-label="Toggle routing settings"
            ><i class="fa fa-cogs" class:text-gray-500={!options.autoRouting}></i></button
        >
    </div>

    {#if showSettings}
        <div in:slide out:slide>
            {#if options.modeOfTransport === "pedestrian" && options.pedestrianOptions}
                <p class="text-sm font-medium pb-1">{$_("walking-speed")}</p>
                <Slider
                    minValue={0.5}
                    maxValue={25}
                    bind:currentValue={options.pedestrianOptions.walking_speed}
                ></Slider>
                <p class="text-sm text-end">
                    {formatSpeed(
                        options.pedestrianOptions.walking_speed! / 3.6,
                    )}
                </p>
                <hr class="border-input-border my-3" />
                <p class="text-sm font-medium">{$_("use-hills")}</p>
                <Slider
                    minValue={0}
                    maxValue={1}
                    step={0.1}
                    bind:currentValue={options.pedestrianOptions!.use_hills}
                ></Slider>
                <p class="text-sm text-end">
                    {options.pedestrianOptions.use_hills?.toFixed(2)}
                </p>
                <hr class="border-input-border my-3" />
                <p class="text-sm font-medium">{$_("max-hiking-difficulty")}</p>
                <Slider
                    minValue={0}
                    maxValue={6}
                    step={1}
                    bind:currentValue={
                        options.pedestrianOptions!.max_hiking_difficulty
                    }
                ></Slider>
                <p class="text-sm text-end">
                    {options.pedestrianOptions.max_hiking_difficulty?.toFixed(
                        0,
                    )}
                </p>
                <hr class="border-input-border my-3" />
                <Toggle
                    label={$_("shortest")}
                    bind:value={options.pedestrianOptions.shortest}
                ></Toggle>
            {:else if options.modeOfTransport === "bicycle" && options.bicycleOptions}
                <Select
                    items={bikeTypes}
                    label={$_("bike-type")}
                    onchange={(v) => adjustSpeeddependingOnBikeType(v)}
                    bind:value={options.bicycleOptions.bicycle_type}
                ></Select>
                <hr class="border-input-border my-3" />
                <p class="text-sm font-medium pb-1">{$_("cycling-speed")}</p>
                <Slider
                    minValue={5}
                    maxValue={50}
                    bind:currentValue={options.bicycleOptions.cycling_speed}
                    bind:this={cycleSpeedSlider}
                ></Slider>
                <p class="text-sm text-end">
                    {formatSpeed(options.bicycleOptions.cycling_speed! / 3.6)}
                </p>
                <hr class="border-input-border my-3" />
                <p class="text-sm font-medium">{$_("use-hills")}</p>
                <Slider
                    minValue={0}
                    maxValue={1}
                    step={0.05}
                    bind:currentValue={options.bicycleOptions.use_hills}
                ></Slider>
                <p class="text-sm text-end">
                    {options.bicycleOptions.use_hills?.toFixed(2)}
                </p>
                <hr class="border-input-border my-3" />
                <p class="text-sm font-medium">{$_("use-roads")}</p>
                <Slider
                    minValue={0}
                    maxValue={1}
                    step={0.05}
                    bind:currentValue={options.bicycleOptions.use_roads}
                ></Slider>
                <p class="text-sm text-end">
                    {options.bicycleOptions.use_roads?.toFixed(2)}
                </p>
                <hr class="border-input-border my-3" />
                <p class="text-sm font-medium">{$_("avoid-bad-surfaces")}</p>
                <Slider
                    minValue={0}
                    maxValue={1}
                    step={0.05}
                    bind:currentValue={
                        options.bicycleOptions.avoid_bad_surfaces
                    }
                ></Slider>
                <p class="text-sm text-end">
                    {options.bicycleOptions.avoid_bad_surfaces?.toFixed(2)}
                </p>
                <hr class="border-input-border my-3" />
                <Toggle
                    label={$_("shortest")}
                    bind:value={options.bicycleOptions.shortest}
                ></Toggle>
            {:else if options.modeOfTransport === "auto" && options.autoOptions}
                <p class="text-sm font-medium pb-1">{$_("fixed-speed")}</p>
                <Slider
                    minValue={0}
                    maxValue={252}
                    step={1}
                    bind:currentValue={options.autoOptions.fixed_speed}
                ></Slider>
                <p class="text-sm text-end">
                    {formatSpeed(options.autoOptions.fixed_speed! / 3.6)}
                </p>
                <hr class="border-input-border my-3" />
                <p class="text-sm font-medium">{$_("top-speed")}</p>
                <Slider
                    minValue={10}
                    maxValue={252}
                    step={1}
                    bind:currentValue={options.autoOptions.top_speed}
                ></Slider>
                <p class="text-sm text-end">
                    {formatSpeed(options.autoOptions.top_speed! / 3.6)}
                </p>
                <hr class="border-input-border my-3" />
                <p class="text-sm font-medium">{$_("car")} {$_("width")}</p>
                <Slider
                    minValue={1}
                    maxValue={10}
                    step={0.1}
                    bind:currentValue={options.autoOptions.width}
                ></Slider>
                <p class="text-sm text-end">
                    {options.autoOptions.width?.toFixed(1)}
                </p>
                <hr class="border-input-border my-3" />
                <p class="text-sm font-medium">{$_("car")} {$_("height")}</p>
                <Slider
                    minValue={1}
                    maxValue={10}
                    step={0.1}
                    bind:currentValue={options.autoOptions.height}
                ></Slider>
                <p class="text-sm text-end">
                    {options.autoOptions.height?.toFixed(1)}
                </p>
                <hr class="border-input-border my-3" />
                <Toggle
                    label={$_("shortest")}
                    bind:value={options.autoOptions.shortest}
                ></Toggle>
            {/if}
        </div>
    {/if}
</div>
