import GPX from "$lib/models/gpx/gpx";
import type Track from "$lib/models/gpx/track";
import TrackSegment from "$lib/models/gpx/track-segment";
import Waypoint from "$lib/models/gpx/waypoint";
import { type RoutingOptions, type ValhallaAnchor, type ValhallaHeightResponse, type ValhallaRouteResponse } from "$lib/models/valhalla";
import { APIError } from "$lib/util/api_util";
import { decodePolyline, encodePolyline } from "$lib/util/polyline_util";
import { get } from "svelte/store";
import { _ } from "svelte-i18n";


const emtpyTrack: Track = { trkseg: [] }

class ValhallaStore {
    route: GPX = $state(new GPX({ trk: [emtpyTrack] }));
    anchors: ValhallaAnchor[] = $state([]);
}

export const valhallaStore = new ValhallaStore();


export function clearRoute() {
    valhallaStore.route = new GPX({ trk: [emtpyTrack] });
}

export function clearAnchors() {
    for (const anchor of valhallaStore.anchors) {
        anchor.marker?.remove();
    }
    valhallaStore.anchors = [];
}

export function setRoute(newRoute: GPX) {
    valhallaStore.route = newRoute
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
        valhallaStore.route.trk?.at(0)?.trkseg?.splice(index, 0, segment);
    } else {
        valhallaStore.route.trk?.at(0)?.trkseg?.push(segment);
    }

    valhallaStore.route.features = valhallaStore.route.getTotals();
}

export async function editRoute(index: number, waypoints: Waypoint[]) {
    const segment = valhallaStore.route.trk?.at(0)?.trkseg?.at(index)
    if (segment) {
        segment.trkpt = waypoints
    }
    valhallaStore.route.features = valhallaStore.route.getTotals();
}

export function deleteFromRoute(index: number) {
    valhallaStore.route.trk?.at(0)?.trkseg?.splice(index, 1);
    valhallaStore.route.features = valhallaStore.route.getTotals();
}

export function reverseRoute() {
    for (const trk of valhallaStore.route.trk ?? []) {
        for (const seg of trk.trkseg ?? []) {
            seg.trkpt?.reverse()
        }
        trk.trkseg?.reverse()
    }
    valhallaStore.route.trk?.reverse()

    valhallaStore.route.features = valhallaStore.route.getTotals();

    valhallaStore.anchors.reverse();

    valhallaStore.anchors.forEach((a, i) => {
        if (!a.marker) {
            return;
        }
        a.marker.getElement().textContent = "" + (i + 1);

        const anchorPopupHeading = a.marker
            .getPopup()
            ._content.getElementsByTagName("h5")[0];
        if (anchorPopupHeading) {
            anchorPopupHeading.textContent =
                get(_)("valhallaStore.route-point") + " #" + (i + 1);
        }
    });
}

export function resetRoute() {
    valhallaStore.route = new GPX({ trk: [{ ...emtpyTrack }] });

    valhallaStore.anchors.forEach((a) => {
        if (!a.marker) {
            return;
        }

        a.marker.remove();
    })

    valhallaStore.anchors = []
}

export async function recalculateHeight() {
    await valhallaStore.route.correctElevation();
}

export function normalizeRouteTime() {
    let currentTime = new Date();

    for (const seg of valhallaStore.route.trk?.at(0)?.trkseg ?? []) {

        if (!seg.trkpt?.length) {
            continue
        }
        const baseTime = seg.trkpt[0].time?.getTime() ?? 0;
        for (let i = 0; i < seg.trkpt.length; i++) {
            const wp = seg.trkpt![i];
            const offset = (wp.time?.getTime() ?? 0) - baseTime;
            const adjustedTime = new Date(currentTime.getTime() + offset);

            wp.time = adjustedTime
        }
        currentTime = new Date(seg.trkpt[seg.trkpt.length - 1].time!.getTime());
    }
}