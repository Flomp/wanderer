import type { LatLng, Marker } from "leaflet";

interface ValhallaRouteResponse {
    trip: {
        legs: {
            shape: string;
        }[];
    };
}

interface ValhallaHeightResponse {
    height: number[];
}

interface ValhallaAnchor {
    id: string,
    lat: number,
    lon: number,
    marker?: Marker
}

export { type ValhallaRouteResponse, type ValhallaHeightResponse, type ValhallaAnchor }