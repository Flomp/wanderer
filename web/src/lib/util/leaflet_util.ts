import type { Waypoint } from "$lib/models/waypoint";
import type { Icon, LeafletEvent, Map, Marker } from "leaflet";

export function createMarkerFromWaypoint(L: any, waypoint: Waypoint, onDragEnd?: (event: LeafletEvent) => void): Marker {
    const fontAwesomeIcon = L.AwesomeMarkers.icon({
        icon: waypoint.icon,
        prefix: "fa",
        markerColor: "cadetblue",
        iconColor: "white",
    }) as Icon;

    const marker = L.marker([waypoint.lat, waypoint.lon], {
        title: waypoint.name,
        icon: fontAwesomeIcon,
        draggable: onDragEnd != null
    })
        .bindPopup(
            "<b>" +
            waypoint.name +
            "</b>" +
            (waypoint.description && waypoint.description.length > 0
                ? "<br>" + waypoint.description
                : ""),
        );
    if (onDragEnd) {
        marker.on("dragend", onDragEnd);
    }

    return marker;
}

export function calculatePixelPerMeter(map: Map, meters: number) {
    const y = map.getSize().y;
    const x = map.getSize().x;
    const maxMeters = map.containerPointToLatLng([0, y]).distanceTo(map.containerPointToLatLng([x, y]));
    const pixelPerMeter = x / maxMeters;

    return pixelPerMeter * meters
}

export function calculateScaleFactor(map: Map) {
    function _pxTOmm() {
        let heightRef = document.createElement('div');
        heightRef.style.height = '1mm';
        heightRef.style.position = "absolute";
        heightRef.id = 'heightRef';
        document.body.appendChild(heightRef);

        const pxPermm = heightRef.getBoundingClientRect().height;

        document.body.removeChild(heightRef);

        return function pxTOmm(px: number) {
            return px / pxPermm;
        }
    }
    var centerOfMap = map.getSize().y / 2;

    var realWorlMetersPer100Pixels = map.distance(
        map.containerPointToLatLng([0, centerOfMap]),
        map.containerPointToLatLng([100, centerOfMap])
    );

    const screenMetersPer100Pixels = _pxTOmm()(100) / 1000;
        
    const scaleFactor = realWorlMetersPer100Pixels / screenMetersPer100Pixels

    return scaleFactor
}