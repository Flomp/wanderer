import Bakery from "$lib/assets/svgs/pois/bakery.svg?raw"
import type { MapMouseEvent, Marker, StyleSpecification } from "maplibre-gl"

export interface BaseLayer {
    markers?: Record<string, Marker>,
    spec: StyleSpecification,
    listener?: {
        onMouseUp?: (e: MapMouseEvent) => void,
        onMouseDown?: (e: MapMouseEvent) => void,
        onEnter?: (e: MapMouseEvent) => void,
        onLeave?: (e: MapMouseEvent) => void,
        onMouseMove?: (e: MapMouseEvent) => void,
    }
}

export const baseMapStyles: Record<string, string | StyleSpecification> = {
    "OpenFreeMap": "/styles/ofm.json",
    "OpenTopoMap": {
        version: 8,
        sources: {
            openTopoMap: {
                type: 'raster',
                tiles: ['https://tile.opentopomap.org/{z}/{x}/{y}.png'],
                tileSize: 256,
                maxzoom: 17,
                attribution:
                    '&copy; <a href="https://www.opentopomap.org" target="_blank">OpenTopoMap</a> &copy; <a href="https://www.openstreetmap.org/copyright" target="_blank">OpenStreetMap</a>',
            },
        },
        layers: [
            {
                id: 'openTopoMap',
                type: 'raster',
                source: 'openTopoMap',
            },
        ],
    },
    "CyclOSM": {
        version: 8,
        sources: {
            cyclOSM: {
                type: 'raster',
                tiles: [
                    'https://a.tile-cyclosm.openstreetmap.fr/cyclosm/{z}/{x}/{y}.png',
                    'https://b.tile-cyclosm.openstreetmap.fr/cyclosm/{z}/{x}/{y}.png',
                    'https://c.tile-cyclosm.openstreetmap.fr/cyclosm/{z}/{x}/{y}.png',
                ],
                tileSize: 256,
                maxzoom: 18,
                attribution:
                    '&copy; <a href="https://github.com/cyclosm/cyclosm-cartocss-style/releases" title="CyclOSM - Open Bicycle render">CyclOSM</a> &copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
            },
        },
        layers: [
            {
                id: 'cyclOSM',
                type: 'raster',
                source: 'cyclOSM',
            },
        ],
    },
    "Carto Light": "https://basemaps.cartocdn.com/gl/positron-gl-style/style.json",
    "Carto Dark": "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json"
}

export const overlays: Record<string, StyleSpecification> = {
    hiking: {
        version: 8,
        name: "waymarkedTrailsHiking",
        sources: {
            waymarkedTrailsHiking: {
                type: 'raster',
                tiles: ['https://tile.waymarkedtrails.org/hiking/{z}/{x}/{y}.png'],
                tileSize: 256,
                maxzoom: 18,
                attribution:
                    '&copy; <a href="https://www.waymarkedtrails.org" target="_blank">Waymarked Trails</a>',
            },
        },
        layers: [
            {
                id: 'waymarkedTrailsHiking',
                type: 'raster',
                source: 'waymarkedTrailsHiking',
            },
        ],
    },
    cycling: {
        version: 8,
        name: "waymarkedTrailsCycling",

        sources: {
            waymarkedTrailsCycling: {
                type: 'raster',
                tiles: ['https://tile.waymarkedtrails.org/cycling/{z}/{x}/{y}.png'],
                tileSize: 256,
                maxzoom: 18,
                attribution:
                    '&copy; <a href="https://www.waymarkedtrails.org" target="_blank">Waymarked Trails</a>',
            },
        },
        layers: [
            {
                id: 'waymarkedTrailsCycling',
                type: 'raster',
                source: 'waymarkedTrailsCycling',
            },
        ],
    },
    MTB: {
        version: 8,
        name: "waymarkedTrailsMTB",

        sources: {
            waymarkedTrailsMTB: {
                type: 'raster',
                tiles: ['https://tile.waymarkedtrails.org/mtb/{z}/{x}/{y}.png'],
                tileSize: 256,
                maxzoom: 18,
                attribution:
                    '&copy; <a href="https://www.waymarkedtrails.org" target="_blank">Waymarked Trails</a>',
            },
        },
        layers: [
            {
                id: 'waymarkedTrailsMTB',
                type: 'raster',
                source: 'waymarkedTrailsMTB',
            },
        ],
    },
    Skiing: {
        name: "waymarkedTrailsWinter",
        version: 8,
        sources: {
            waymarkedTrailsWinter: {
                type: 'raster',
                tiles: ['https://tile.waymarkedtrails.org/slopes/{z}/{x}/{y}.png'],
                tileSize: 256,
                maxzoom: 18,
                attribution:
                    '&copy; <a href="https://www.waymarkedtrails.org" target="_blank">Waymarked Trails</a>',
            },
        },
        layers: [
            {
                id: 'waymarkedTrailsWinter',
                type: 'raster',
                source: 'waymarkedTrailsWinter',
            },
        ],
    }
}

export type POI = { q: string, icon: { svg: any, bg: string } }
export const pois: Record<string, POI> = {
    "grocery-store": { q: "nwr[shop=supermarket];nwr[shop=convenience];", icon: { svg: "", bg: "red" } },
    "bakery": { q: "nwr[shop=bakery];", icon: { svg: Bakery, bg: "coral" } },
    "food-drinks": { q: "nwr[amenity=restaurant];nwr[amenity=fast_food];nwr[amenity=cafe];nwr[amenity=pub];nwr[amenity=bar];", icon: { svg: "", bg: "red" } },
    "toilet": { q: "nwr[amenity=toilets];", icon: { svg: "", bg: "red" } },
    "drinking-water": { q: "nwr[amenity=drinking_water];nwr[amenity=water_point];nwr[natural=spring][drinking_water=yes];", icon: { svg: "", bg: "red" } },
    "shower": { q: "nwr[amenity=shower];", icon: { svg: "", bg: "red" } },
    "shelter": { q: "nwr[amenity=shelter];", icon: { svg: "", bg: "red" } },
    "barrier": { q: "nwr[barrier=true];", icon: { svg: "", bg: "red" } },
    "attraction": { q: "nwr[tourism=attraction];", icon: { svg: "", bg: "red" } },
    "viewpoint": { q: "nwr[tourism=viewpoint];", icon: { svg: "", bg: "red" } },
    "hotel": { q: "nwr[tourism=hotel];nwr[tourism=hostel];nwr[tourism=guest_house];nwr[tourism=motel];", icon: { svg: "", bg: "red" } },
    "camp-site": { q: "nwr[tourism=camp_site];", icon: { svg: "", bg: "red" } },
    "hut": { q: "nwr[tourism=alpine_hut];nwr[tourism=wilderness_hut];", icon: { svg: "", bg: "red" } },
    "peak": { q: "nwr[natural=peak];", icon: { svg: "", bg: "red" } },
    "mountain-pass": { q: "nwr[mountain_pass=yes];", icon: { svg: "", bg: "red" } },
    "climbing": { q: "nwr[sport=climbing];", icon: { svg: "", bg: "red" } },
    "bicylce-parking": { q: "nwr[amenity=bicycle_parking];", icon: { svg: "", bg: "red" } },
    "bicycle-rental": { q: "nwr[amenity=bicycle_rental];", icon: { svg: "", bg: "red" } },
    "bicycle-shop": { q: "nwr[shop=bicycle];", icon: { svg: "", bg: "red" } },
    "gas-station": { q: "nwr[amenity=fuel];", icon: { svg: "", bg: "red" } },
    "parking": { q: "nwr[amenity=parking];", icon: { svg: "", bg: "red" } },
    "car-repair": { q: "nwr[shop=car_repair];", icon: { svg: "", bg: "red" } },
    "motorcycle-repair": { q: "nwr[shop=motorcycle_repair];", icon: { svg: "", bg: "red" } },
    "railway-station": { q: "nwr[railway=station];", icon: { svg: "", bg: "red" } },
    "subway": { q: "nwr[railway=subway_entrance];", icon: { svg: "", bg: "red" } },
    "tram": { q: "nwr[railway=tram_stop];", icon: { svg: "", bg: "red" } },
    "bus": { q: "nwr[public_transport=stop_position][bus=yes];nwr[public_transport=platform][bus=yes];", icon: { svg: "", bg: "red" } },
    "ferry": { q: "nwr[amenity=ferry_terminal];", icon: { svg: "", bg: "red" } },
}



export type MapState = {
    base: keyof typeof baseMapStyles,
    overlays: { [K in keyof typeof overlays]: boolean }
    pois: Record<string, Record<string, boolean>>
}

export const defaultMapState: MapState = {
    base: "OpenFreeMap",
    overlays: {
        waymarkedTrailsHiking: false,
        waymarkedTrailsCycling: false,
        waymarkedTrailsHorseRiding: false,
        waymarkedTrailsMTB: false,
        waymarkedTrailsSkating: false,
        waymarkedTrailsWinter: false,
    },
    pois: {
        food: {
            "food-drinks": false,
            "grocery-store": false,
            bakery: false,
        },
        tourism: {
            "camp-site": false,
            attraction: false,
            hotel: false,
            hut: false,
            viewpoint: false,
        },
        ammenity: {
            "drinking-water": false,
            barrier: false,
            shelter: false,
            shower: false,
            toilet: false,
        },
        hiking: {
            "mountain-pass": false,
            climbing: false,
            peak: false,
        },
        cycling: {
            "bicycle-rental": false,
            "bicycle-shop": false,
            "bicylce-parking": false,
        },
        "car-motorcycle": {
            "car-repair": false,
            "gas-station": false,
            "motorcycle-repair": false,
            parking: false,
        },
        "public-transport": {
            "railway-station": false,
            bus: false,
            ferry: false,
            subway: false,
            tram: false,
        }
    }
}