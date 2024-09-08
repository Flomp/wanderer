<script lang="ts">
    import { page } from "$app/stores";
    import { Settings } from "$lib/models/settings";
    import type { Trail } from "$lib/models/trail";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import type AutoGraticule from "$lib/vendor/leaflet-graticule/leaflet-auto-graticule";
    import type { Layer, Map, Marker, Polyline } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import { createEventDispatcher, onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import Dropdown from "../base/dropdown.svelte";

    export let trails: Trail[];
    export let markers: Marker[] = [];
    export let map: Map | null = null;
    export let options: any = {};
    export let graticule: AutoGraticule | null = null;
    export let crosshair: boolean = false;
    export let activeTrailIndex: number = 0;

    const dispatch = createEventDispatcher();

    let L: any;
    let gpxGroup: any;

    let selectedMetric: "altitude" | "slope" | "speed" | false = "altitude";

    $: gpxData = trails.map((t) => t.expand.gpx_data);
    $: if (gpxData && gpxGroup) {
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

        const AutoGraticule = (
            await import("$lib/vendor/leaflet-graticule/leaflet-auto-graticule")
        ).default;

        map = L.map("map", {
            preferCanvas: true,
            plugins: ["/vendor/leaflet-elevation/libs/leaflet-gpxgroup.js"],
        }).setView(
            [
                trails.at(activeTrailIndex)?.lat ?? 0,
                trails.at(activeTrailIndex)?.lon ?? 0,
            ],
            3,
        );
        map!.attributionControl.setPrefix(false);

        map!.on("zoomend", function () {
            dispatch("zoomend", map);
        });

        map!.on("click", function (e) {
            dispatch("click", e);
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
            ...($page.data.settings as Settings)?.tilesets?.reduce<
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

        map!.on("baselayerchange", function (e) {
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
            imperial: $page.data.settings?.unit == "imperial" ?? false,
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
            // points_options: opts.points,
            elevation: true,
            elevation_options: elevation_options,
            flyToBounds: true,
            distanceMarkers: true,
        });

        const markerLayerGroup = L.layerGroup().addTo(map);

        gpxGroup.on("selection_changed", ({ polyline }: { polyline: any }) => {
            markerLayerGroup.clearLayers();

            activeTrailIndex = polyline.options.index ?? 0;

            if (polyline._selected) {
                for (const waypoint of trails.at(activeTrailIndex)?.expand
                    .waypoints ?? []) {
                    const marker = createMarkerFromWaypoint(L, waypoint);
                    marker.addTo(markerLayerGroup!);
                    markers.push(marker);
                }
            }
        });

        gpxGroup.addTo(map);
        // controlElevation = L.control.elevation(elevation_options).addTo(map);

        if (elevation_options.graticule) {
            graticule = new AutoGraticule();
            graticule.addTo(map!);
        }
    });

    function switchHotline(metric: "altitude" | "slope" | "speed" | false) {
        gpxGroup._elevation.updateOptions({
            hotline: metric,
            autofitBounds: false,
        });
        selectedMetric = metric;
        localStorage.setItem("gradient", metric.toString());
        gpxGroup.clear();
        gpxGroup._tracks = gpxData;
        gpxGroup.addTracks();
    }
</script>

<div id="map-container" class="flex flex-col h-full">
    <div
        id="map"
        class="rounded-xl z-0 basis-full min-h-96 md:min-h-0"
        style={crosshair
            ? "position: relative; outline-style: none;cursor: crosshair !important"
            : "position: relative; outline-style: none;"}
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
    <div class="flex items-center justify-between">
        <slot />
        <div class="basis-[300px] flex-grow flex-shrink-0" id="elevation"></div>
    </div>
</div>
