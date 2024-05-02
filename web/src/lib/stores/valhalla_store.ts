import GPX from "$lib/models/gpx/gpx";
import type Track from "$lib/models/gpx/track";
import TrackSegment from "$lib/models/gpx/track-segment";
import Waypoint from "$lib/models/gpx/waypoint";
import { type ValhallaAnchor, type ValhallaHeightResponse, type ValhallaRouteResponse } from "$lib/models/valhalla";
import { decodePolyline, encodePolyline } from "$lib/util/polyline_util";
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

export async function calculateRouteBetween(startLat: number, startLon: number, endLat: number, endLon: number, costing: string = "pedestrian", autoRoute: boolean = true) {

    let shape;
    if (autoRoute) {
        let costingBody;
        switch (costing) {
            case "bicycle":
                costingBody = { "costing": "bicycle", "costing_options": { "bicycle": { "bicycle_type": "Hybrid", "use_roads": 0.5, "use_hills": 0.5, "avoid_bad_surfaces": 0.5, "use_ferry": 0 } } }
                break;
            case "auto":
                costingBody = { "costing": "auto", "costing_options": { "auto": { "use_ferry": 0 } } }
            default:
                costingBody = { "costing": costing, "costing_options": { costing: { "max_hiking_difficulty": 6, "use_ferry": 0 } } }
                break;
        }
        const requestBody = {
            "directions_type": "none",
            "locations": [{ "lat": startLat, "lon": startLon }, { "lat": endLat, "lon": endLon }],
            ...costingBody
        }

        let r = await fetch("/api/v1/valhalla/route", { method: "POST", body: JSON.stringify(requestBody) })

        if (!r.ok) {
            throw new ClientResponseError(await r.json())
        }
        const routeResponse: ValhallaRouteResponse = await r.json();
        shape = routeResponse.trip.legs[0].shape
    } else {
        shape = encodePolyline([[startLat, startLon], [endLat, endLon]])
    }

    const r2 = await fetch("/api/v1/valhalla/height", { method: "POST", body: JSON.stringify({ encoded_polyline: shape }) })

    if (!r2.ok) {
        throw new ClientResponseError(await r2.json())
    }
    const heightResponse: ValhallaHeightResponse = await r2.json()
    const points = decodePolyline(shape);
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