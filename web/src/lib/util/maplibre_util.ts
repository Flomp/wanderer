import type { Trail } from "$lib/models/trail";
import type { Waypoint } from "$lib/models/waypoint";
import M from "maplibre-gl";
import { getFileURL } from "./file_util";
import { formatDistance, formatElevation, formatTimeHHMM } from "./format_util";
import { get } from "svelte/store";
import { _ } from "svelte-i18n";
import { haversineDistance } from "$lib/models/gpx/utils";

export class FontawesomeMarker extends M.Marker {
    constructor(options: { icon: string, fontSize?: string, width?: number, backgroundColor?: string, fontColor?: string, id?: string }, markerOptions?: M.MarkerOptions) {
        const element = document.createElement('div')
        element.className = `cursor-pointer flex items-center justify-center w-${options.width ?? 7} aspect-square bg-${options.backgroundColor ?? "gray-500"} rounded-full text-${options.fontSize ?? "normal"}`
        element.id = options.id ?? "";
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
        draggable: onDragEnd !== undefined,
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

export function createPopupFromTrail(trail: Trail) {
    const thumbnail = trail.photos.length
        ? getFileURL(trail, trail.photos[trail.thumbnail])
        : "/imgs/default_thumbnail.webp";
    const popup = new M.Popup({maxWidth: "320px"});
    popup.setHTML(
        `<a href="/map/trail/${trail.id}" data-sveltekit-preload-data="off">
    <li class="flex items-center gap-4 cursor-pointer text-black max-w-80">
        <div class="shrink-0"><img class="h-14 w-14 object-cover rounded-xl" src="${thumbnail}" alt="">
        </div>
        <div>
            <h4 class="font-semibold text-lg">${trail.name}</h4>
            <div class="flex gap-x-4">
            ${trail.location ? `<h5><i class="fa fa-location-dot mr-2"></i>${trail.location}</h5>` : ""}
            <h5><i class="fa fa-gauge mr-2"></i>${get(_)(trail.difficulty as string)}</h5>
            </div>
            <div class="grid grid-cols-2 mt-2 gap-x-4 gap-y-2 text-sm text-gray-500 flex-wrap"><span class="shrink-0"><i
                        class="fa fa-left-right mr-2"></i>${formatDistance(
            trail.distance,
        )}</span><span class="shrink-0"><i class="fa fa-clock mr-2"></i>${formatTimeHHMM(
            trail.duration,
        )}</span><span class="shrink-0"><i class="fa fa-arrow-trend-up mr-2"></i>${formatElevation(
            trail.elevation_gain,
        )}</span></span> <span class="shrink-0"><i class="fa fa-arrow-trend-down mr-2"></i>${formatElevation(
            trail.elevation_loss,
        )}</span></div>
        </div>
    </li>
</a>`)
    return popup;
}

export function calculatePixelPerMeter(map: M.Map, meters: number) {
    const y = map.getCanvas().getBoundingClientRect().y;
    const x = map.getCanvas().getBoundingClientRect().x;
    const maxMeters = map.unproject([0, y]).distanceTo(map.unproject([x, y]));
    const pixelPerMeter = x / maxMeters;

    return pixelPerMeter * meters
}

export function calculateScaleFactor(map: M.Map) {
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
    var centerOfMap = map.getCanvas().getBoundingClientRect().y / 2;

    const p1 = map.unproject([0, centerOfMap]);
    const p2 = map.unproject([100, centerOfMap]);
    var realWorldMetersPer100Pixels = haversineDistance(
        p1.lat, p1.lng, p2.lat, p2.lng
    );

    const screenMetersPer100Pixels = _pxTOmm()(100) / 1000;

    const scaleFactor = realWorldMetersPer100Pixels / screenMetersPer100Pixels

    return scaleFactor
}