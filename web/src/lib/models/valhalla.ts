import * as M from "maplibre-gl";

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
    marker?: M.Marker
}

export { type ValhallaAnchor, type ValhallaHeightResponse, type ValhallaRouteResponse };
