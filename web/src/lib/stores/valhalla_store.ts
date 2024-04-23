import GPX from "$lib/models/gpx/gpx";
import type Track from "$lib/models/gpx/track";
import TrackSegment from "$lib/models/gpx/track-segment";
import Waypoint from "$lib/models/gpx/waypoint";
import { type ValhallaAnchor, type ValhallaHeightResponse, type ValhallaRouteResponse } from "$lib/models/valhalla";
import { ClientResponseError } from "pocketbase";


const emtpyTrack: Track = { trkseg: [] }
export let route: GPX = new GPX({ trk: [emtpyTrack] });
export let anchors: ValhallaAnchor[] = [];

export function clearRoute() {
    route = new GPX({ trk: [emtpyTrack] });
    anchors = [];
}

export function setRoute(newRoute: GPX) {
    route = newRoute
}

export async function calculateRouteBetween(startLat: number, startLon: number, endLat: number, endLon: number) {
    const requestBody = {
        "directions_type": "none",
        "locations": [{ "lat": startLat, "lon": startLon }, { "lat": endLat, "lon": endLon }],
        "costing": "pedestrian", "costing_options": { "pedestrian": { "max_hiking_difficulty": 6, "use_ferry": 0 } }
    }
    let r = await fetch("/api/v1/valhalla/route", { method: "POST", body: JSON.stringify(requestBody) })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
    const routeResponse: ValhallaRouteResponse = await r.json();
    const shape = routeResponse.trip.legs[0].shape
    r = await fetch("/api/v1/valhalla/height", { method: "POST", body: JSON.stringify({ encoded_polyline: shape }) })

    if (!r.ok) {
        throw new ClientResponseError(await r.json())
    }
    const heightResponse: ValhallaHeightResponse = await r.json()
    const points = decodeShape(shape);
    const waypoints = points.map((p, i) => new Waypoint({ $: { lat: p[0], lon: p[1] }, ele: heightResponse.height[i] }))

    return waypoints
}

export async function appendToRoute(waypoints: Waypoint[]) {
    const segment = new TrackSegment({ trkpt: [] })

    for (const wpt of waypoints) {
        segment.trkpt!.push(wpt)
    }
    route.trk?.at(0)?.trkseg?.push(segment);
}

export async function editRoute(index: number, waypoints: Waypoint[]) {
    const segment = route.trk?.at(0)?.trkseg?.at(index)
    if (segment) {
        segment.trkpt = waypoints
    }
}

export function deleteFromRoute(index: number) {
    route.trk?.at(0)?.trkseg?.splice(index, 1);
}


function decodeShape(shape: string, precision: number = 6) {
    var index = 0,
        lat = 0,
        lng = 0,
        coordinates = [],
        shift = 0,
        result = 0,
        byte = null,
        latitude_change,
        longitude_change,
        factor = Math.pow(10, precision);

    while (index < shape.length) {
        byte = null;
        shift = 0;
        result = 0;

        do {
            byte = shape.charCodeAt(index++) - 63;
            result |= (byte & 0x1f) << shift;
            shift += 5;
        } while (byte >= 0x20);

        latitude_change = ((result & 1) ? ~(result >> 1) : (result >> 1));

        shift = result = 0;

        do {
            byte = shape.charCodeAt(index++) - 63;
            result |= (byte & 0x1f) << shift;
            shift += 5;
        } while (byte >= 0x20);

        longitude_change = ((result & 1) ? ~(result >> 1) : (result >> 1));

        lat += latitude_change;
        lng += longitude_change;

        coordinates.push([lat / factor, lng / factor]);
    }

    return coordinates;
};