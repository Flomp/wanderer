import type { LngLatBounds, StyleSpecification } from "maplibre-gl";
import { pois, type BaseLayer, type MapState } from "./layers";
import type { OverpassResponse } from "./types";

export class OverpassLayer implements BaseLayer {
    private overpassApiURL: string = "https://overpass.private.coffee/api/interpreter"

    data: GeoJSON.FeatureCollection = ({ type: 'FeatureCollection', features: [] });
    pois: string[] = [];

    spec: StyleSpecification = {
        version: 8,
        name: "overpass",
        sources: {
            overpass: {
                type: 'geojson',
                data: this.data,
            }
        },
        layers: [
            {
                id: 'overpass',
                type: 'symbol',
                source: 'overpass',
                layout: {
                    'icon-image': ['get', 'icon'],
                    'icon-size': 0.25,
                    'icon-padding': 0,
                    'icon-allow-overlap': ['step', ['zoom'], false, 14, true],
                },
            }
        ],
    };


    async updateLayer(state: MapState, bounds: LngLatBounds) {
        let activePOIs: string[] = []
        for (const category of Object.keys(state.pois)) {
            for (const [name, active] of Object.entries(state.pois[category])) {
                if (active) {
                    activePOIs.push(name)
                }
            }
        }
        const q = this.getOverpassQuery(activePOIs, bounds)
        if (!q.length) {
            return;
        }
        const r = await fetch(`${this.overpassApiURL}?data=${q}`)
        const response: OverpassResponse = await r.json();

        this.updateData(response)
        this.pois = activePOIs
    }

    private updateData(data: OverpassResponse, ) {
        let pois: GeoJSON.Feature[] = [];

        if (data.elements === undefined) {
            return;
        }

        for (let element of data.elements) {
            console.log(element);
            
            pois.push({
                type: 'Feature',
                geometry: {
                    type: 'Point',
                    coordinates: element.center
                        ? [element.center.lon, element.center.lat]
                        : [element.lon, element.lat],
                },
                properties: {
                    id: element.id,
                    lat: element.center ? element.center.lat : element.lat,
                    lon: element.center ? element.center.lon : element.lon,
                    icon: `overpass-bakery`,
                    tags: element.tags,
                    type: element.type,
                },
            });
        }
        this.data.features = pois;
    }

    private getOverpassQuery(data: string[], bounds: LngLatBounds) {
        const q = data.map(p => pois[p].q).join('')
        return `[bbox:${bounds.getSouth()},${bounds.getWest()},${bounds.getNorth()},${bounds.getEast()}][out:json];(${q});out center;`;
    }

}