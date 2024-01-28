import type { Waypoint } from "$lib/models/waypoint";
import type { Icon, Marker } from "leaflet";

export function createMarkerFromWaypoint(L: any, waypoint: Waypoint): Marker {
    const fontAwesomeIcon = L.AwesomeMarkers.icon({
        icon: waypoint.icon,
        prefix: "fa",
        markerColor: "cadetblue",
        iconColor: "white",
    }) as Icon;

    const marker = L.marker([waypoint.lat, waypoint.lon], {
        title: waypoint.name,
        icon: fontAwesomeIcon,
    })
        .bindPopup(
            "<b>" +
            waypoint.name +
            "</b>" +
            (waypoint.description && waypoint.description.length > 0
                ? "<br>" + waypoint.description
                : ""),
        );

    return marker;
}