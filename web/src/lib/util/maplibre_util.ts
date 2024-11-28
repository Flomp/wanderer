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

export function createAnchorMarker(lat: number, lon: number, index: number, onDeleteClick: () => void, onDragEnd: (event: M.Marker) => void): FontawesomeMarker {

    const anchorElement = document.createElement("span")
    anchorElement.className = "cursor-pointer rounded-full w-6 h-6 border border-black text-center bg-background-inverse text-content-inverse"
    anchorElement.textContent = "" + index
    const marker = new M.Marker(
        {
            draggable: true,
            element: anchorElement
        }
    );
    marker.setLngLat([lon, lat]);

    const deleteButton = document.createElement("button");
    deleteButton.className = "fa fa-trash text-red-500 rounded-full aspect-square h-8 text-lg";
    deleteButton.addEventListener("click", onDeleteClick)
    const popup = new M.Popup({})
    popup.setDOMContent(deleteButton)
    marker.setPopup(popup);

    marker.on("dragend", onDragEnd);
    marker.getElement().addEventListener("click", (e) => {
        e.stopPropagation();
        marker.togglePopup();
    })

    return marker
}

// export function calculatePixelPerMeter(map: Map, meters: number) {
//     const y = map.getSize().y;
//     const x = map.getSize().x;
//     const maxMeters = map.containerPointToLatLng([0, y]).distanceTo(map.containerPointToLatLng([x, y]));
//     const pixelPerMeter = x / maxMeters;

//     return pixelPerMeter * meters
// }

// export function calculateScaleFactor(map: Map) {
//     function _pxTOmm() {
//         let heightRef = document.createElement('div');
//         heightRef.style.height = '1mm';
//         heightRef.style.position = "absolute";
//         heightRef.id = 'heightRef';
//         document.body.appendChild(heightRef);

//         const pxPermm = heightRef.getBoundingClientRect().height;

//         document.body.removeChild(heightRef);

//         return function pxTOmm(px: number) {
//             return px / pxPermm;
//         }
//     }
//     var centerOfMap = map.getSize().y / 2;

//     var realWorlMetersPer100Pixels = map.distance(
//         map.containerPointToLatLng([0, centerOfMap]),
//         map.containerPointToLatLng([100, centerOfMap])
//     );

//     const screenMetersPer100Pixels = _pxTOmm()(100) / 1000;

//     const scaleFactor = realWorlMetersPer100Pixels / screenMetersPer100Pixels

//     return scaleFactor
// }