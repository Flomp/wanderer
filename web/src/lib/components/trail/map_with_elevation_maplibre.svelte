<script lang="ts">
    import { page } from "$app/stores";
    import type { Trail } from "$lib/models/trail";
    import { theme } from "$lib/stores/theme_store";
    import { findStartAndEndPoints } from "$lib/util/geojson_util";
    import { toGeoJson } from "$lib/util/gpx_util";
    import {
        createMarkerFromWaypoint,
        FontawesomeMarker,
    } from "$lib/util/maplibre_util";
    import type { ElevationProfileControl } from "$lib/vendor/maplibre-elevation-profile/elevationprofile-control";
    import type { GeoJsonObject } from "geojson";
    import * as M from "maplibre-gl";
    import "maplibre-gl/dist/maplibre-gl.css";
    import { onDestroy, onMount } from "svelte";

    export let trail: Trail;
    export let markers: M.Marker[] = [];
    export let map: M.Map | null = null;
    export let crosshairCursor: boolean = false;

    let mapContainer: HTMLDivElement;

    let epc: ElevationProfileControl;

    $: data = toGeoJson(trail.expand.gpx_data!) as GeoJsonObject;

    $: if ($theme == "dark") {
        epc?.toggleTheme({
            profileBackgroundColor: "#191b24",
            elevationGridColor: "#ddd2",
            labelColor: "#ddd8",
            crosshairColor: "#fff5",
        });
    } else {
        epc?.toggleTheme({
            profileBackgroundColor: "#242734",
            elevationGridColor: "#0002",
            labelColor: "#0009",
            crosshairColor: "#0005",
        });
    }

    onMount(async () => {
        const initialState = {
            lng: 0,
            lat: 0,
            zoom: 14,
        };
        const ElevationProfileControl = (
            await import(
                "$lib/vendor/maplibre-elevation-profile/elevationprofile-control"
            )
        ).ElevationProfileControl;

        map = new M.Map({
            container: mapContainer,
            style: `https://basemaps.cartocdn.com/gl/positron-gl-style/style.json`,
            center: [initialState.lng, initialState.lat],
            zoom: initialState.zoom,
        });

        for (const waypoint of trail?.expand.waypoints ?? []) {
            const marker = createMarkerFromWaypoint(waypoint);
            marker.addTo(map);
            markers.push(marker);
        }

        const startEndPoint = findStartAndEndPoints(data);

        const startMarker = new FontawesomeMarker(
            { icon: "fa fa-bullseye" },
            {},
        )
            .setLngLat(startEndPoint[0] as M.LngLatLike)
            .addTo(map);

        const endMarker = new FontawesomeMarker(
            { icon: "fa fa-flag-checkered" },
            {},
        )
            .setLngLat(startEndPoint[1] as M.LngLatLike)
            .addTo(map);

        const elevationMarker = new FontawesomeMarker(
            {
                icon: "fa-regular fa-circle",
                fontSize: "xs",
                width: 4,
                backgroundColor: "primary",
                fontColor: "white",
            },
            {},
        );
        elevationMarker.setLngLat([0, 0]).addTo(map);
        elevationMarker.setOpacity("0");

        epc = new ElevationProfileControl({
            visible: true,
            profileBackgroundColor: $theme == "light" ? "#242734" : "#191b24",
            backgroundColor: "bg-menu-background/90",
            unit: $page.data.settings?.unit ?? "metric",
            profileLineWidth: 3,
            displayDistanceGrid: true,
            tooltipDisplayDPlus: false,
            onEnter: () => {
                elevationMarker.setOpacity("1");
            },
            onLeave: () => {
                elevationMarker.setOpacity("0");
            },
            onMove: (data) => {
                elevationMarker.setLngLat(data.position as M.LngLatLike);
            },
        });
        map.addControl(new M.NavigationControl());
        map.addControl(
            new M.ScaleControl({
                maxWidth: 120,
                unit: $page.data.settings?.unit ?? "metric",
            }),
            "top-left",
        );
        map.addControl(epc);
        epc.setData(data, trail.expand.waypoints);

        map.on("load", () => {
            map!.addSource("trail-source", {
                type: "geojson",
                data: data as any,
            });

            map!.addLayer({
                id: "uploaded-polygons",
                type: "line",
                source: "trail-source",
                paint: {
                    "line-color": "#648ad5",
                    "line-width": 5,
                },
            });
            map!.fitBounds(data.bbox as any, {
                animate: false,
                padding: {
                    top: 16,
                    left: 16,
                    right: 16,
                    bottom: map!.getContainer().clientHeight * 0.3 + 16,
                },
            });
        });
    });

    onDestroy(() => {
        map?.remove();
    });
</script>

<div id="map" class:cursor-pointer={crosshairCursor} bind:this={mapContainer}></div>

<style>
    #map {
        width: 100%;
        height: 100%;
    }
</style>
