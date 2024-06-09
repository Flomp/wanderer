<script lang="ts">
    import { page } from "$app/stores";
    import type { Trail } from "$lib/models/trail";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import "$lib/vendor/leaflet-elevation/src/index.css";
    import type AutoGraticule from "$lib/vendor/leaflet-graticule/leaflet-auto-graticule";
    import type { Map, Marker } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import { createEventDispatcher, onMount } from "svelte";

    export let trail: Trail | null;
    export let markers: Marker[] = [];
    export let map: Map | null = null;
    export let options = {};
    export let graticule: AutoGraticule | null = null;
    export let crosshair: boolean = false;

    const dispatch = createEventDispatcher();

    let L: any;
    let controlElevation: any;

    $: gpxData = trail?.expand.gpx_data;
    $: if (gpxData && controlElevation) {       
        controlElevation.clear();
        controlElevation.load(gpxData);
    }

    $: if(options) {        
        controlElevation?.updateOptions(options)    
    }

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");
        //@ts-ignore
        await import("$lib/vendor/leaflet-elevation/src/index.js");
        const AutoGraticule = (
            await import("$lib/vendor/leaflet-graticule/leaflet-auto-graticule")
        ).default;

        map = L.map("map", { preferCanvas: true }).setView(
            [trail?.lat ?? 0, trail?.lon ?? 0],
            3,
        );
        map!.attributionControl.setPrefix(false);

        map!.on("zoomend", function () {
            dispatch("zoomend", map);
        });

        map!.on("click", function (e) {
            dispatch("click", e);
        });

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);

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
            distanceMarkers: { lazy: false, distance: false, direction: true, offset: 500 },
            // Toggle "leaflet-edgescale" integration
            edgeScale: false,
            // Toggle "leaflet-hotline" integration
            hotline: true,
            // Display track datetimes: true || false
            timestamps: false,
            waypoints: false,
            wptIcons: false,
            wptLabels: false,
            preferCanvas: true,
            trkStart: {
                interactive: false,
                className: "hihi",
                icon: L.AwesomeMarkers.icon({
                    icon: "circle-half-stroke",
                    prefix: "fa",
                    markerColor: "cadetblue",
                    iconColor: "white",
                    className: "awesome-marker pointer-events-none"
                }),
            },
            trkEnd: {
                icon: L.AwesomeMarkers.icon({
                    icon: "flag-checkered",
                    prefix: "fa",
                    markerColor: "cadetblue",
                    iconColor: "white",
                    className: "awesome-marker pointer-events-none"
                }),
            },
            graticule: false,
            drawing: false
        };

        const elevation_options = Object.assign(
            default_elevation_options,
            options,
        );

        controlElevation = L.control.elevation(elevation_options).addTo(map);

        for (const waypoint of trail?.expand.waypoints ?? []) {
            const marker = createMarkerFromWaypoint(L, waypoint);
            marker.addTo(map!);
            markers.push(marker);
        }

        if (elevation_options.graticule) {
            graticule = new AutoGraticule();
            graticule.addTo(map!);
        }
    });
</script>

<div id="map-container" class="flex flex-col h-full">
    <div
        id="map"
        class="rounded-xl z-0 basis-full min-h-72"
        style={crosshair
            ? "position: relative; outline-style: none;cursor: crosshair !important"
            : "position: relative; outline-style: none;"}
    ></div>
    <div class="flex items-center justify-between">
        <slot />
        <div class="basis-[300px] flex-grow flex-shrink-0" id="elevation"></div>
    </div>
</div>
