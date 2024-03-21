<script lang="ts">
    import type { Trail } from "$lib/models/trail";
    import { currentUser } from "$lib/stores/user_store";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import type { Map, Marker } from "leaflet";
    import { onMount } from "svelte";

    export let trail: Trail;
    export let markers: Marker[] = [];

    let L: any;
    let map: Map;
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

        map = L.map("map").setView([trail.lat ?? 0, trail.lon ?? 0], 14);
        map.attributionControl.setPrefix(false)

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);

        const elevation_options = {
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

        controlElevation = L.control.elevation(elevation_options).addTo(map);

        for (const waypoint of trail.expand.waypoints) {
            const marker = createMarkerFromWaypoint(L, waypoint);
            marker.addTo(map);
            markers.push(marker);
        }
    });
</script>

<div id="map-container" class="flex flex-col">
    <div id="map" class="rounded-xl z-0 basis-full"></div>
    <div id="elevation"></div>
</div>

<style>
    #map-container {
        height: calc(100vh - 180px);
    }
    @media only screen and (min-width: 768px) {
        #map-container {
            height: calc(100vh - 124px);
        }
    }
</style>
