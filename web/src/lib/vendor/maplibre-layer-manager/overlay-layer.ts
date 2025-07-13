import type { StyleSpecification } from "maplibre-gl";
import { overlays, type BaseLayer } from "./layers";

export class OverlayLayer implements BaseLayer {
    spec: StyleSpecification;

    constructor(name: string) {
        this.spec = overlays[name]
    }
}