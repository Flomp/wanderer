import type { Waypoint } from "$lib/models/waypoint";
import type { Icon, LeafletEvent, Marker } from "leaflet";

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