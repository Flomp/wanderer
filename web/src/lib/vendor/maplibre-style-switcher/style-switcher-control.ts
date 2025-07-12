import type { RadioItem } from "$lib/components/base/radio_group.svelte";
import StyleSwitcher from "$lib/components/map/style_switcher.svelte";
import { type ControlPosition, type IControl } from "maplibre-gl";
import { mount } from "svelte";
/**
 * Style switcher control options
 */

export type StyleSwitcherControlOptions = {
    styles: RadioItem[];
    layers: RadioItem[];
    onMapStyleSwitch: (index: number, style: RadioItem) => void
    onLayerChange: (checked: boolean, layer: RadioItem) => void

    state: {
        selectedStyle: number;
        selectedLayers: Record<string, boolean>
    };
};

export class StyleSwitcherControl implements IControl {
    private container?: HTMLElement;

    private switcherContainer?: HTMLDivElement;
    private settings: StyleSwitcherControlOptions;

    constructor(options: StyleSwitcherControlOptions) {
        if (typeof window === "undefined")
            throw new Error("This pluggin must be mounted client-side");
        this.settings = { ...options };
    }

    getContainer(): HTMLDivElement | undefined {
        return this.switcherContainer;
    }

    onAdd(): HTMLElement {
        this.container = document.createElement('div');
        mount(StyleSwitcher, { target: this.container, props: { settings: this.settings } });
        return this.container;
    }

    onRemove(): void {
        // remove button
        if (this.container?.parentNode) {
            this.container.parentNode.removeChild(this.container);
        }
    }

    getDefaultPosition?: (() => ControlPosition) | undefined;
}