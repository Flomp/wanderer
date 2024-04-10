<script lang="ts">
    import type { Trail } from "$lib/models/trail";
    import { currentUser } from "$lib/stores/user_store";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import type { Map, Marker } from "leaflet";
    import { createEventDispatcher, onMount } from "svelte";

    export let trail: Trail;
    export let markers: Marker[] = [];
    export let map: Map | null = null;
    export let options = {};

    const dispatch = createEventDispatcher();

    let L: any;
    let controlElevation: any;

    $: if (trail.expand.gpx_data && controlElevation) {
        controlElevation.load(trail.expand.gpx_data);
    }

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");
        //@ts-ignore
        await import("$lib/vendor/leaflet-elevation/src/index.js");

        map = L.map("map", { preferCanvas: true }).setView(
            [trail.lat ?? 0, trail.lon ?? 0],
            14,
        );
        map!.attributionControl.setPrefix(false);

        map!.on("zoomend", function () {
            dispatch("zoomend", map);
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
            imperial: $currentUser?.unit == "imperial" ?? false,
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
            almostOver: true,
            // Toggle "leaflet-distance-markers" integration
            distanceMarkers: false,
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
                icon: L.AwesomeMarkers.icon({
                    icon: "circle-half-stroke",
                    prefix: "fa",
                    markerColor: "cadetblue",
                    iconColor: "white",
                }),
            },
            trkEnd: {
                icon: L.AwesomeMarkers.icon({
                    icon: "flag-checkered",
                    prefix: "fa",
                    markerColor: "cadetblue",
                    iconColor: "white",
                }),
            },
        };

        const elevation_options = Object.assign(
            default_elevation_options,
            options,
        );

        controlElevation = L.control.elevation(elevation_options).addTo(map);

        for (const waypoint of trail.expand.waypoints) {
            const marker = createMarkerFromWaypoint(L, waypoint);
            marker.addTo(map!);
            markers.push(marker);
        }
    });
</script>

<div id="map-container" class="flex flex-col">
    <div id="map" class="rounded-xl z-0 basis-full"></div>
    <div class="flex items-center justify-between">
        <slot />
        <div class="basis-full min-w-[300px]" id="elevation"></div>
    </div>
</div>
