import type { StyleSpecification } from "maplibre-gl";
import type { BaseLayer } from "./layers";

export class PreviewLayer implements BaseLayer {

    spec: StyleSpecification;

    constructor(geojson: GeoJSON.FeatureCollection) {
        const startPoints: GeoJSON.FeatureCollection = {
            type: "FeatureCollection",
            features: geojson.features.map((f, i) => ({
                type: "Feature",
                properties: {
                    ...f.properties,
                    id: i
                },
                geometry: {
                    type: "Point",
                    coordinates: (f.geometry as any).coordinates[0]
                }
            }))
        };
        this.spec = {
            version: 8,
            name: "preview",
            sources: {
                "preview": {
                    type: "geojson",
                    data: geojson,
                },
                "preview-start-point": {
                    type: "geojson",
                    data: startPoints,
                }
            },
            layers: [
                {
                    id: "preview",
                    type: "line",
                    source: "preview",
                    minzoom: 10,
                    paint: {
                        "line-color": ["get", "color"],
                        "line-width": 5,
                    },
                },
                {
                    id: "preview-start-point",
                    type: "circle",
                    source: "preview-start-point",
                    minzoom: 10,
                    paint: {
                        "circle-color": "#242734",
                        "circle-radius": 6,
                        "circle-stroke-width": 2,
                        "circle-stroke-color": "#fff",
                    },
                }
            ]

        };
    }
}