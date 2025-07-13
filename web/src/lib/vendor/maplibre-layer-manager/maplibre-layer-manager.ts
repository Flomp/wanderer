import * as M from "maplibre-gl";
import { baseMapStyles, defaultMapState, pois, type BaseLayer, type MapState } from "./layers";
import { OverlayLayer } from "./overlay-layer";
import { OverpassLayer } from "./overpass-layer";



export class LayerManager {
    private map: M.Map;
    state!: MapState;
    layers: Record<string, BaseLayer> = {};

    constructor(map: M.Map) {
        this.map = map;

        const storedMapState = localStorage.getItem("map-state")
        if (storedMapState) {
            this.state = JSON.parse(storedMapState)
        } else {
            this.state = defaultMapState
        }
    }

    init() {
        this.map.on('moveend', () => { });

        this.map.on('styledata', () => {
            this.restoreLayers();
        })
        try {
            this.update(this.state, true);

            const overpassLayer = new OverpassLayer()
            this.addLayer("overpass", overpassLayer)

        } catch (e) {
            console.error(e)
            // map is probably not initialized yet
        }
    }

    update(newState: MapState, initialize: boolean = false) {
        const oldState = this.state;

        if (oldState.base != newState.base || initialize) {
            this.updateBaseLayer(baseMapStyles[newState.base])
        }

        for (const [name, active] of Object.entries(newState.overlays)) {
            const oldOverlayActive = oldState.overlays[name]

            if (active && (!oldOverlayActive || initialize)) {
                this.addLayer(name, new OverlayLayer(name))
            } else if (oldOverlayActive && !active) {
                this.removeLayer(name)
            }
        }

        const overpassLayer = this.layers.overpass;
        if (overpassLayer) {
            const castedOverpassLayer = overpassLayer as OverpassLayer
            castedOverpassLayer.updateLayer(newState, this.map.getBounds()).then(() => {
                this.loadIcons(castedOverpassLayer.pois);
                (this.map.getSource('overpass') as M.GeoJSONSource).setData(castedOverpassLayer.data);
            })
        }

        this.state = newState
        localStorage.setItem("map-state", JSON.stringify(this.state));
    }

    private updateBaseLayer(layer: string | M.StyleSpecification) {
        this.map.setStyle(layer);
    }


    private addLayer(id: string, layer: BaseLayer) {
        for (const [id, s] of Object.entries(layer.spec.sources)) {
            if (!this.map.getSource(id)) {
                this.map.addSource(id, s)
            }
        }

        for (const l of layer.spec.layers) {
            if (!this.map.getLayer(l.id)) {
                this.map.addLayer(l)
            }
        }

        this.layers[id] = layer
    }

    private removeLayer(id: string) {
        const layer = this.layers[id];
        if (!layer) {
            return;
        }
        for (const l of layer.spec.layers) {
            if (this.map.getLayer(l.id)) {
                this.map.removeLayer(l.id)
            }
        }

        for (const [id, _] of Object.entries(layer.spec.sources)) {
            if (this.map.getSource(id)) {
                this.map.removeSource(id)
            }
        }
        delete this.layers[id]

    }

    private restoreLayers() {
        for (const [id, layer] of Object.entries(this.layers)) {
            this.addLayer(id, layer)
        }
    }

    private loadIcons(activePois: string[]) {
        activePois.forEach((poi) => {
            if (!this.map.hasImage(`overpass-${poi}`)) {
                let icon = new Image(100, 100);
                icon.onload = () => {
                    if (!this.map.hasImage(`overpass-${poi}`)) {
                        this.map.addImage(`overpass-${poi}`, icon);
                    }
                };

                const svg = `
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 40 40">
                    <circle cx="20" cy="20" r="20" fill="${pois[poi].icon.bg}" />
                    <g transform="translate(8 8) scale(0.05)">
                    ${pois[poi].icon.svg}
                    </g>
                </svg>
                `
                
                icon.src =
                    'data:image/svg+xml,' +
                    encodeURIComponent(svg);
            }
        });
    }
}