import type { Waypoint } from "$lib/models/waypoint";
import M from "maplibre-gl";

export class FontawesomeMarker extends M.Marker {
    constructor(options: { icon: string, fontSize?: string, width?: number, backgroundColor?: string, fontColor?: string }, markerOptions?: M.MarkerOptions) {
        const element = document.createElement('div')
        element.className = `cursor-pointer flex items-center justify-center w-${options.width ?? 7} aspect-square bg-${options.backgroundColor ?? "gray-500"} rounded-full text-${options.fontSize ?? "normal"}`
        super({ element: element, ...markerOptions });

        const {
            icon,
        } = options

        const iconElementString = `<i class="text-${options.fontColor ?? "white"} ${icon}"></i>`
        this._element.insertAdjacentHTML('beforeend', iconElementString)
    }
}

export function createMarkerFromWaypoint(waypoint: Waypoint, onDragEnd?: (marker: M.Marker) => void): FontawesomeMarker {
    const marker = new FontawesomeMarker({
        icon: `fa fa-${waypoint.icon}`,
    }, {
        draggable: onDragEnd != null,
        color: "#6b7280"

    })
    const popup = new M.Popup({ offset: 25, closeButton: false }).setHTML(
        "<b>" +
        waypoint.name +
        "</b>" +
        (waypoint.description && waypoint.description.length > 0
            ? "<br>" + waypoint.description
            : ""),
    );
    marker
        .setLngLat([waypoint.lon, waypoint.lat])
        .setPopup(popup)

    if (onDragEnd) {
        marker.on("dragend", () => onDragEnd(marker));
    }

    return marker;
}