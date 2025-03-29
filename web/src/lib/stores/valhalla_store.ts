import GPX from "$lib/models/gpx/gpx";
import type Track from "$lib/models/gpx/track";
import TrackSegment from "$lib/models/gpx/track-segment";
import Waypoint from "$lib/models/gpx/waypoint";
import { type RoutingOptions, type ValhallaAnchor, type ValhallaHeightResponse, type ValhallaRouteResponse } from "$lib/models/valhalla";
import { APIError } from "$lib/util/api_util";
import { decodePolyline, encodePolyline } from "$lib/util/polyline_util";


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

export async function calculateRouteBetween(startLat: number, startLon: number, endLat: number, endLon: number, options: RoutingOptions) {

    let shape;
    let duration: number;
    if (options.autoRouting) {
        let costingBody;
        switch (options.modeOfTransport) {
            case "bicycle":
                costingBody = { "costing": options.modeOfTransport, "costing_options": { [options.modeOfTransport]: options.autoOptions } }
                break;
            case "auto":
                costingBody = { "costing": options.modeOfTransport, "costing_options": { [options.modeOfTransport]: options.bicycleOptions } }
                break;

            case "pedestrian":
                costingBody = { "costing": options.modeOfTransport, "costing_options": { [options.modeOfTransport]: options.pedestrianOptions } }
                break;
        }
        const requestBody = {
            "directions_type": "none",
            "locations": [{ "lat": startLat, "lon": startLon }, { "lat": endLat, "lon": endLon }],
            ...costingBody
        }

        let r = await fetch("/api/v1/valhalla/route", { method: "POST", body: JSON.stringify(requestBody) })

        if (!r.ok) {
            const response = await r.json();
            throw new APIError(r.status, response.message, response.detail)
        }

        const routeResponse: ValhallaRouteResponse = await r.json();
        shape = routeResponse.trip.legs[0].shape
        duration = routeResponse.trip.summary.time
    } else {
        shape = encodePolyline([[startLat, startLon], [endLat, endLon]])
        duration = 0;
    }

    const r2 = await fetch("/api/v1/valhalla/height", { method: "POST", body: JSON.stringify({ encoded_polyline: shape }) })

    if (!r2.ok) {
        const response = await r2.json();
        throw new APIError(r2.status, response.message, response.detail)
    }

    const heightResponse: ValhallaHeightResponse = await r2.json()
    const points = decodePolyline(shape);
    const startTime = new Date().getTime();

    const waypoints = points.map((p, i) => new Waypoint({ $: { lat: p[0], lon: p[1] }, ele: heightResponse.height[i], time: new Date(startTime + (((duration * 1000) / points.length) * i)) }))

    return waypoints
}

export async function insertIntoRoute(waypoints: Waypoint[], index?: number) {
    const segment = new TrackSegment({ trkpt: waypoints })

    if (index) {
        route.trk?.at(0)?.trkseg?.splice(index, 0, segment);
    } else {
        route.trk?.at(0)?.trkseg?.push(segment);
    }

    route.features = route.getTotals();
}

export async function editRoute(index: number, waypoints: Waypoint[]) {
    const segment = route.trk?.at(0)?.trkseg?.at(index)
    if (segment) {
        segment.trkpt = waypoints
    }
    route.features = route.getTotals();

}

export function deleteFromRoute(index: number) {
    route.trk?.at(0)?.trkseg?.splice(index, 1);
    route.features = route.getTotals();

}