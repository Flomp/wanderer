import * as M from "maplibre-gl";
import { type ControlPosition, type IControl } from "maplibre-gl";
/**
 * Style switcher control options
 */

type MapStyle = {
    text: string;
    value: string;
    thumbnail?: string;
}

export type StyleSwitcherControlOptions = {
    styles: MapStyle[];
    onSwitch: (style: MapStyle) => void
    selectedIndex: number;
};

export class StyleSwitcherControl implements IControl {
    private buttonContainer?: HTMLDivElement;
    private toggleButton?: HTMLButtonElement;
    private isSwitcherShown = false;
    private iconSpan?: HTMLSpanElement;
    private styleLis: HTMLLIElement[] = [];

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

    onAdd(map: M.Map): HTMLElement {
        this.buttonContainer = document.createElement("div");

        this.buttonContainer.classList.add(
            "maplibregl-ctrl",
            "maplibregl-ctrl-group",
            "relative",
            "z-10"
        );
        this.toggleButton = document.createElement("button");
        this.buttonContainer.appendChild(this.toggleButton);
        // this.buttonContainer.classList.add("w-16", "aspect-square")
        this.iconSpan = document.createElement("i");
        this.iconSpan.classList.add("fa", "fa-layer-group", "text-black");
        this.toggleButton.appendChild(this.iconSpan);
        this.toggleButton.addEventListener("click", this.toggleSwitcher.bind(this));


        this.switcherContainer = document.createElement("div");
        this.switcherContainer.style.setProperty("display", "none");
        this.switcherContainer.classList.add("bg-menu-background", "rounded-lg", "py-3", "mt-2", "shadow-xl")
        this.switcherContainer.style.setProperty("position", "absolute");
        this.switcherContainer.style.setProperty("transform", "translateX(calc(-100% + 29px))");
        // this.switcherContainer.style.setProperty("max-width", "120px");

        const headingDiv = document.createElement("div");
        headingDiv.classList.add("flex", "items-center", "gap-x-12", "px-3");
        const closeButton = document.createElement("button");
        closeButton.classList.add("fa", "fa-close", "!rounded-full");
        closeButton.addEventListener("click", () => this.hideSwitcher());
        const switcherContainerHeading = document.createElement("span")
        switcherContainerHeading.classList.add("text-lg", "font-semibold", "whitespace-nowrap")
        switcherContainerHeading.textContent = "Select map style"
        headingDiv.appendChild(switcherContainerHeading);
        headingDiv.appendChild(closeButton);

        this.switcherContainer.appendChild(headingDiv)

        const buttonDiv = document.createElement("ul")
        buttonDiv.classList.add("mt-2", "max-h-64", "overflow-y-scroll");
        this.settings.styles.forEach((style, i) => {
            const styleLi = document.createElement("li");
            styleLi.classList.add("flex", "items-center", "gap-x-4", "px-3", "py-2", "cursor-pointer", "hover:bg-menu-item-background-hover")
            if (style.thumbnail) {
                const styleImg = document.createElement("img");
                styleImg.classList.add("w-12", "h-12", "rounded-md")
                styleImg.src = style.thumbnail;
                styleLi.appendChild(styleImg)
            } else {
                const styleIconContainer = document.createElement("i");
                styleIconContainer.classList.add("w-12", "h-12", "rounded-md", "bg-blue-200", "flex", "items-center", "justify-center")
                const styleIcon = document.createElement("i");
                styleIcon.classList.add("fa", "fa-map-location-dot", "text-xl", "text-black")
                styleIconContainer.appendChild(styleIcon)
                styleLi.appendChild(styleIconContainer)
            }

            const styleName = document.createElement("span");
            styleName.classList.add("!text-sm")
            if (i == this.settings.selectedIndex) {
                styleName.classList.add("font-semibold")
            }
            styleName.textContent = style.text;
            styleLi.appendChild(styleName)

            styleLi.addEventListener("click", () => {
                this.hideSwitcher();
                this.styleLis?.at(this.settings.selectedIndex)?.lastElementChild?.classList.remove("font-semibold")
                this.settings.selectedIndex = i;
                styleName.classList.add("font-semibold")
                this.settings.onSwitch(style)
            })

            this.styleLis?.push(styleLi)
            buttonDiv.appendChild(styleLi)
        })
        this.switcherContainer.appendChild(buttonDiv)
        this.buttonContainer.appendChild(this.switcherContainer);


        return this.buttonContainer;
    }

    private toggleSwitcher() {
        if (!this.switcherContainer) return;

        if (this.isSwitcherShown) {
            this.hideSwitcher();
        } else {
            this.showSwitcher();
        }
    }

    showSwitcher() {
        this.switcherContainer?.style.setProperty("display", "inherit");
        this.isSwitcherShown = true;
    }

    hideSwitcher() {
        this.switcherContainer?.style.setProperty("display", "none");
        this.isSwitcherShown = false;
    }

    onRemove(): void {
        // remove button
        if (this.buttonContainer?.parentNode) {
            this.buttonContainer.parentNode.removeChild(this.buttonContainer);
        }
        this.buttonContainer = undefined;
        this.toggleButton = undefined;
        this.isSwitcherShown = false;
    }

    getDefaultPosition?: (() => ControlPosition) | undefined;
}