<script lang="ts">
    import { page } from "$app/stores";
    import { Settings } from "$lib/models/settings";
    import type { Trail } from "$lib/models/trail";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import { createEventDispatcher, onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import Dropdown from "../base/dropdown.svelte";

    export let trails: Trail[];
    export let map: any | null = null;
    export let options: any = {};
    export let activeTrailIndex: number | null = 0;
    export let markers: any[] = [];
    export let bindRoutePopup: boolean = true;

    const dispatch = createEventDispatcher();

    let L: any;
    let gpxGroup: any;

    let selectedMetric: "altitude" | "slope" | "speed" | false = "altitude";

    $: gpxData = trails.map((t) => {
        return { id: t.id, gpx: t.expand.gpx_data };
    });
    $: if (
        gpxData &&
        gpxGroup &&
        map &&
        (gpxData.length != gpxGroup._tracks.length ||
            gpxGroup?._tracks[0] !== gpxData[0])
    ) {
        // if (gpxData.length == 0) {
        //     map?.setView([0, 0], 4);
        // }
        gpxGroup._elevation.updateOptions({
            autofitBounds: options.autofitBounds ?? true,
        });
        gpxGroup.clear();
        gpxGroup._tracks = gpxData;
        gpxGroup.addTracks();
    }

    $: if (options && gpxGroup) {
        gpxGroup._elevation.updateOptions(options);
    }

    $: hotlineSwitcherItems = [
        {
            text: $_("altitude"),
            value: "altitude",
            icon: selectedMetric == "altitude" ? "circle-dot" : "circle",
        },
        {
            text: $_("slope"),
            value: "slope",
            icon: selectedMetric == "slope" ? "circle-dot" : "circle",
        },
        {
            text: $_("speed"),
            value: "speed",
            icon: selectedMetric == "speed" ? "circle-dot" : "circle",
        },
        {
            text: $_("off"),
            value: false,
            icon: selectedMetric == false ? "circle-dot" : "circle",
        },
    ];

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");
        await import("$lib/vendor/leaflet-elevation/src/index.js");
        await import("$lib/vendor/leaflet-elevation/libs/leaflet-gpxgroup");

        map = L.map("map", {
            preferCanvas: true,
            plugins: ["/vendor/leaflet-elevation/libs/leaflet-gpxgroup.js"],
        }).setView(
            [
                trails.at(activeTrailIndex ?? 0)?.lat ?? 0,
                trails.at(activeTrailIndex ?? 0)?.lon ?? 0,
            ],
            3,
        );

        dispatch("init", map);
        map!.attributionControl.setPrefix(false);

        map!.on("zoomend", function () {
            dispatch("zoomend", map);
        });

        map!.on("click", function (e: any) {
            dispatch("click", e);
        });

        map!.on("moveend", function (e: any) {
            dispatch("moveend", e);
        });

        const baseLayer = L.tileLayer(
            "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
            {
                attribution: "© OpenStreetMap contributors",
            },
        );

        const topoLayer = L.tileLayer(
            "https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png",
            {
                attribution: "© OpenStreetMap contributors",
            },
        );

        const baseMaps: Record<string, L.TileLayer> = {
            OpenStreetMaps: baseLayer,
            OpenTopoMaps: topoLayer,
            ...($page.data?.settings as Settings)?.tilesets?.reduce<
                Record<string, string>
            >((t, current) => {
                t[current.name] = L.tileLayer(current.url);
                return t;
            }, {}),
        };

        L.control.layers(baseMaps).addTo(map);

        const layerPreference = localStorage.getItem("layer");

        if (
            layerPreference &&
            Object.keys(baseMaps).includes(layerPreference)
        ) {
            baseMaps[layerPreference].addTo(map!);
        } else {
            baseLayer.addTo(map);
        }

        map!.on("baselayerchange", function (e: any) {
            localStorage.setItem("layer", e.name);
        });

        const localMetric =
            (localStorage.getItem("gradient") as any) ?? "altitude";
        selectedMetric = localMetric === "false" ? false : localMetric;

        const default_elevation_options = {
            height: 200,

            theme: "lightblue-theme",
            detached: true,
            elevationDiv: "#elevation",
            closeBtn: false,
            followMarker: true,
            autofitBounds: true,
            imperial: $page.data.settings?.unit == "imperial",
            reverseCoords: false,
            acceleration: false,
            slope: true,
            speed: true,
            altitude: true,
            time: true,
            distance: true,
            // Summary track info style: "inline" || "multiline" || false
            summary: false,
            downloadLink: false,
            ruler: false,
            legend: true,
            // Toggle "leaflet-almostover" integration
            almostOver: false,
            // Toggle "leaflet-distance-markers" integration
            distanceMarkers: {
                lazy: false,
                distance: false,
                direction: true,
                offset: 1000,
            },
            // Toggle "leaflet-edgescale" integration
            edgeScale: false,
            // Toggle "leaflet-hotline" integration
            hotline: selectedMetric,
            // Display track datetimes: true || false
            timestamps: false,
            waypoints: false,
            wptIcons: false,
            wptLabels: false,
            preferCanvas: true,
            graticule: false,
            drawing: false,
        };

        const elevation_options = Object.assign(
            default_elevation_options,
            options,
        );

        gpxGroup = L.gpxGroup(gpxData, {
            points: [],
            elevation: true,
            elevation_options: elevation_options,
            flyToBounds: options.flyToBounds,
            distanceMarkers: true,
            itinerary: options.itinerary,
        });

        const markerLayerGroup = L.layerGroup().addTo(map);

        gpxGroup.on("selection_changed", ({ polyline }: { polyline: any }) => {
            markerLayerGroup.clearLayers();
            activeTrailIndex = null;
            markers = [];

            if (polyline && polyline._selected) {
                activeTrailIndex = trails.findIndex(
                    (t) => t.id == polyline.options.id,
                );

                for (const waypoint of trails.at(activeTrailIndex!)?.expand
                    .waypoints ?? []) {
                    const marker = createMarkerFromWaypoint(L, waypoint);
                    marker.addTo(markerLayerGroup!);
                    markers.push(marker);
                }

                dispatch(
                    "select",
                    trails.find((t) => t.id == polyline.options.id),
                );
            }
        });

        gpxGroup.on("clear", () => {
            markerLayerGroup.clearLayers();
            activeTrailIndex = null;
        });

        gpxGroup.on("route_loaded", ({ route }: { route: any }) => {
            const trail = trails.find((t) => t.id == route.options.id);

            if (!trail) {
                return;
            }
            const thumbnail = trail.photos.length
                ? getFileURL(trail, trail.photos[trail.thumbnail])
                : "/imgs/default_thumbnail.webp";
            if (bindRoutePopup) {
                route.bindPopup(
                    `<a href="/trail/view/${trail.id}" data-sveltekit-preload-data="off">
    <li class="flex items-center gap-4 cursor-pointer text-black max-w-80">
        <div class="shrink-0"><img class="h-14 w-14 object-cover rounded-xl" src="${thumbnail}" alt="">
        </div>
        <div>
            <h4 class="font-semibold text-lg">${trail.name}</h4>
            <div class="flex gap-x-4">
            ${trail.location ? `<h5><i class="fa fa-location-dot mr-2"></i>${trail.location}</h5>` : ""}
            <h5><i class="fa fa-gauge mr-2"></i>${$_(trail.difficulty as string)}</h5>
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
</a>`,
                );
            }
        });

        gpxGroup.addTo(map);
    });

    export function highlightTrail(id: string) {
        gpxGroup?.highlightTrack(id);
    }

    export function unHighlightTrail(id: string) {
        gpxGroup?.unHighlightTrack(id);
    }

    export function selectTrail(id: string) {
        gpxGroup?.select(id);
    }

    export function resetSelection() {
        gpxGroup?.resetSelection();
    }

    export function openPopup(id: string) {
        gpxGroup?.openPopup(id);
    }

    export function closePopup(id: string) {
        gpxGroup?.closePopup(id);
    }

    function switchHotline(metric: "altitude" | "slope" | "speed" | false) {
        if (metric === selectedMetric) {
            return;
        }
        gpxGroup._elevation.updateOptions({
            hotline: metric,
        });
        gpxGroup.options.flyToBounds = false;
        selectedMetric = metric;
        localStorage.setItem("gradient", metric.toString());
        gpxGroup.clear();
        gpxGroup._tracks = gpxData;
        gpxGroup.addTracks();

        gpxGroup.options.flyToBounds = true;
    }
</script>

<div id="map-container" class="h-full relative">
    <div
        id="map"
        class="rounded-xl z-0 h-full min-h-96 md:min-h-0"
        style="position: relative; outline-style: none;"
    >
        <div class="absolute top-20 right-3 text-sm" style="z-index: 500">
            <Dropdown
                items={hotlineSwitcherItems}
                on:change={(e) => switchHotline(e.detail.value)}
                let:toggleMenu={openDropdown}
            >
                <button
                    class="rounded-md border-2 border-black border-opacity-30 bg-white text-black hover:bg-gray-200 focus:ring-4 ring-gray-100/50 transition-colors h-12 w-12"
                    on:click={openDropdown}
                >
                    <i class="fa fa-route text-lg"></i>
                </button>
            </Dropdown>
        </div>
    </div>
    <div class="absolute bottom-0 bg-background/80 w-full" id="elevation"></div>
</div>
