import type { StyleSpecification } from "maplibre-gl";
import type { BaseLayer } from "./layers";

export class DebugLayer implements BaseLayer {
    spec: StyleSpecification = {
        version: 8,
        name: "debug",
        sources: {
            debug: {
                type: 'geojson',
                data: {
                    type: 'FeatureCollection',
                    features: [],
                },
            }
        },
        layers: [{
            id: 'debug-layer',
            type: 'line',
            source: 'debug',
            paint: {
                'line-color': '#ff0000',
                'line-width': 2,
                'line-dasharray': [2, 2],
            },
        }]

    };
}