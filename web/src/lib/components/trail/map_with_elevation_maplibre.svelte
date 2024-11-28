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
    import { StyleSwitcherControl } from "$lib/vendor/maplibre-style-switcher/style-switcher-control";
    import type { GeoJSON } from "geojson";
    import * as M from "maplibre-gl";
    import "maplibre-gl/dist/maplibre-gl.css";
    import { createEventDispatcher, onDestroy, onMount } from "svelte";

    export let trail: Trail | null;
    export let markers: M.Marker[] = [];
    export let map: M.Map | null = null;
    export let drawing: boolean = false;

    let mapContainer: HTMLDivElement;
    let epc: ElevationProfileControl;
    let startMarker: M.Marker;
    let endMarker: M.Marker;

    const dispatch = createEventDispatcher();

    $: data = trail?.expand.gpx_data
        ? (toGeoJson(trail.expand.gpx_data!) as GeoJSON)
        : null;

    $: if (data && map) {
        initMap();
    }

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

    $: if (drawing && map) {
        map.getCanvas().style.cursor = "crosshair";
    } else if (!drawing && map) {
        map.getCanvas().style.cursor = "inherit";
        addStartEndMarkers();
    }

    function initMap() {
        if (!map || !data) {
            return;
        }

        epc.setData(data, trail!.expand.waypoints);
        epc.showProfile();

        const trailSource = map.getSource("trail-source");
        if (!trailSource) {
            addTrailLayer();
        } else {
            (trailSource as M.GeoJSONSource).setData(data);
        }

        if (!drawing) {
            addStartEndMarkers();
        }
    }

    function addTrailLayer() {
        if (!data || !map) {
            return;
        }
        const trailSource = map.getSource("trail-source");
        if (!trailSource) {
            try {
                map.addSource("trail-source", {
                    type: "geojson",
                    data: data,
                });
            } catch (e) {
                return;
            }
        }

        const trailLayer = map.getLayer("trail-layer");
        if (!trailLayer) {
            map.addLayer({
                id: "trail-layer",
                type: "line",
                source: "trail-source",
                paint: {
                    "line-color": "#648ad5",
                    "line-width": 5,
                },
            });
        }
    }

    function addStartEndMarkers() {
        if (!map || !data) {
            return;
        }
        const startEndPoint = findStartAndEndPoints(data);

        startMarker ??= new FontawesomeMarker({ icon: "fa fa-bullseye" }, {});

        startMarker.setLngLat(startEndPoint[0] as M.LngLatLike).addTo(map);

        endMarker ??= new FontawesomeMarker(
            { icon: "fa fa-flag-checkered" },
            {},
        );
        endMarker.setLngLat(startEndPoint[1] as M.LngLatLike).addTo(map);

        map!.fitBounds(data.bbox as any, {
            animate: false,
            padding: {
                top: 16,
                left: 16,
                right: 16,
                bottom: map!.getContainer().clientHeight * 0.3 + 16,
            },
        });
    }

    onMount(async () => {
        const initialState = {
            lng: 0,
            lat: 0,
            zoom: 1,
        };
        const ElevationProfileControl = (
            await import(
                "$lib/vendor/maplibre-elevation-profile/elevationprofile-control"
            )
        ).ElevationProfileControl;

        const mapStyles = [
            {
                text: "Carto Light",
                value: "https://basemaps.cartocdn.com/gl/positron-gl-style/style.json",
                thumbnail:
                    "https://basemaps.cartocdn.com/light_all/1/0/0@2x.png",
            },
            {
                text: "Carto Dark",
                value: "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json",
                thumbnail:
                    "https://basemaps.cartocdn.com/dark_all/1/0/0@2x.png",
            },
        ];
        const preferredMapStyleIndex = mapStyles.findIndex(
            (s) => s.text === localStorage.getItem("layer"),
        );

        map = new M.Map({
            container: mapContainer,
            style:
                mapStyles[preferredMapStyleIndex].value ?? mapStyles[0].value,
            center: [initialState.lng, initialState.lat],
            zoom: initialState.zoom,
        });

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
            visible: false,
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

        const switcherControl = new StyleSwitcherControl({
            styles: mapStyles,
            onSwitch: (style) => {
                map?.setStyle(style.value);
                localStorage.setItem("layer", style.text);
            },
            selectedIndex:
                preferredMapStyleIndex !== -1 ? preferredMapStyleIndex : 0,
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
        map.addControl(switcherControl);

        map.on("styledata", () => {
            addTrailLayer();
        });

        map.on("click", (e) => {
            dispatch("click", e);
        });

        for (const waypoint of trail?.expand.waypoints ?? []) {
            const marker = createMarkerFromWaypoint(waypoint);
            marker.addTo(map);
            markers.push(marker);
        }
    });

    onDestroy(() => {
        map?.remove();
    });
</script>

<div id="map" bind:this={mapContainer}></div>

<style>
    #map {
        width: 100%;
        height: 100%;
    }
</style>
