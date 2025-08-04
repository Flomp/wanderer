import type { StyleSpecification } from "maplibre-gl";
import type { BaseLayer } from "./layers";

export class TerrainLayer implements BaseLayer {
    spec: StyleSpecification;

    constructor(terrainURL: string, hillshadingURL?: string) {
        this.spec = {
            version: 8,
            name: "terrain",
            sources: {
                terrain: {
                    type: "raster-dem",
                    url: terrainURL,
                },
                ...(hillshadingURL ? {
                    hillshading: {
                        type: "raster-dem",
                        url: hillshadingURL
                    }
                } : {})

            },
            layers: [{
                id: "hillshading",
                source: "terrain",
                type: "hillshade",
            }]

        }

    }
}