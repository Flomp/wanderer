<script lang="ts">
    import { page } from "$app/stores";
    import type { Trail } from "$lib/models/trail";
    import { theme } from "$lib/stores/theme_store";
    import { findStartAndEndPoints } from "$lib/util/geojson_util";
    import { toGeoJson } from "$lib/util/gpx_util";
    import {
        createMarkerFromWaypoint,
        createPopupFromTrail,
        FontawesomeMarker,
    } from "$lib/util/maplibre_util";
    import type { ElevationProfileControl } from "$lib/vendor/maplibre-elevation-profile/elevationprofile-control";
    import { StyleSwitcherControl } from "$lib/vendor/maplibre-style-switcher/style-switcher-control";
    import type { GeoJSON } from "geojson";
    import * as M from "maplibre-gl";
    import "maplibre-gl/dist/maplibre-gl.css";
    import { createEventDispatcher, onDestroy, onMount } from "svelte";
    import MaplibreGraticule from "$lib/vendor/maplibre-graticule/maplibre-graticule";
    import { FullscreenControl } from "$lib/vendor/maplibre-fullscreen/fullscreen-control";
    import type { Settings } from "$lib/models/settings";

    export let trails: Trail[] = [];
    export let markers: M.Marker[] = [];
    export let map: M.Map | null = null;
    export let drawing: boolean = false;
    export let showElevation: boolean = true;
    export let showInfoPopup: boolean = false;
    export let showGrid: boolean = false;
    export let showStyleSwitcher: boolean = true;
    export let showFullscreen: boolean = false;
    export let showTerrain: boolean = false;
    export let fitBounds: "animate" | "instant" | "off" = "instant";

    export let elevationProfileContainer: string | HTMLDivElement | undefined =
        undefined;
    export let mapOptions: Partial<M.MapOptions> | undefined = undefined;

    export let activeTrail: number = 0;
    export let minZoom: number = 0;

    let mapContainer: HTMLDivElement;
    let epc: ElevationProfileControl | null = null;
    let graticule: MaplibreGraticule | null = null;

    let layers: Record<
        string,
        {
            startMarker: M.Marker | null;
            endMarker: M.Marker | null;
            source: M.GeoJSONSource | null;
            layer: M.LineLayerSpecification | null;
            listener: {
                onEnter: ((e: M.MapMouseEvent) => void) | null;
                onLeave: ((e: M.MapMouseEvent) => void) | null;
                onClick: ((e: M.MapMouseEvent) => void) | null;
            };
        }
    > = {};

    const dispatch = createEventDispatcher();

    $: data = getData(trails);

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
        addStartEndMarkers(
            trails[activeTrail],
            trails[activeTrail]?.id ?? activeTrail.toString(),
            data?.at(activeTrail),
        );
    }

    $: if (showGrid) {
        if (!graticule) {
            graticule = new MaplibreGraticule({
                minZoom: 0,
                maxZoom: 20,
                showLabels: true,
                labelType: "hdms",
                labelSize: 10,
                labelColor: "#858585",
                longitudePosition: "top",
                latitudePosition: "right",
                paint: {
                    "line-opacity": 0.8,
                    "line-color": "rgba(0,0,0,0.2)",
                },
            });
        }
        map?.addControl(graticule);
    } else {
        if (graticule) {
            map?.removeControl(graticule);
        }
    }

    function getData(trails: Trail[]) {
        if (!trails.length) {
            return [];
        }

        const r: GeoJSON[] = [];
        trails.forEach((t) => {
            if (t.expand.gpx_data) {
                r.push(toGeoJson(t.expand.gpx_data) as GeoJSON);
            } else if (t.lat && t.lon) {
                r.push({
                    id: "",
                    type: "Feature",
                    properties: {},
                    geometry: {
                        type: "Point",
                        coordinates: [t.lon ?? 0, t.lat ?? 0],
                    },
                } as GeoJSON);
            }
        });
        return r;
    }

    function initMap() {
        if (!map) {
            return;
        }

        if (data[activeTrail] && showElevation) {
            epc?.setData(
                data[activeTrail]!,
                trails.at(activeTrail)!.expand.waypoints,
            );
            epc?.showProfile();
        }

        trails.forEach((t, i) => {
            const layerId = t.id ?? i.toString();
            addTrailLayer(t, layerId, data[i]);
        });

        Object.keys(layers).forEach((layerId) => {
            const isStillVisible = trails.some((t) => t.id === layerId);
            if (!isStillVisible) {
                removeTrailLayer(layerId);
            }
        });

        if (
            !drawing &&
            fitBounds !== "off" &&
            data.some((d) => d.bbox !== undefined)
        ) {
            flyToBounds(fitBounds == "animate");
        }
    }

    function getBounds() {
        let minX = Infinity,
            minY = Infinity,
            maxX = -Infinity,
            maxY = -Infinity;

        for (const [xMin, yMin, xMax, yMax] of data
            .filter((d) => d.bbox !== undefined)
            .map((d) => d.bbox!)) {
            minX = Math.min(minX, xMin);
            minY = Math.min(minY, yMin);
            maxX = Math.max(maxX, xMax);
            maxY = Math.max(maxY, yMax);
        }

        return new M.LngLatBounds([minX, minY, maxX, maxY]);
    }

    function flyToBounds(animate: boolean = true) {
        const bounds = data[activeTrail]
            ? (data[activeTrail].bbox as M.LngLatBoundsLike)
            : getBounds();

        if (!bounds) {
            return;
        }

        map!.fitBounds(bounds, {
            animate: animate,
            padding: {
                top: 16,
                left: 16,
                right: 16,
                bottom:
                    16 +
                    (epc?.isProfileShown
                        ? map!.getContainer().clientHeight * 0.3
                        : 0),
            },
        });
    }

    function removeTrailLayer(id: string) {
        if (!layers[id]) {
            return;
        }
        const layer = layers[id];

        if (layer.layer) {
            map?.removeLayer(id);
        }
        if (layer.source) {
            map?.removeSource(id);
        }
        layer.startMarker?.remove();
        layer.endMarker?.remove();

        map?.off("mouseenter", id, layers[id].listener.onEnter!);
        map?.off("mouseleave", id, layers[id].listener.onLeave!);
        map?.off("click", id, layers[id].listener.onClick!);

        delete layers[id];
    }

    function createEmptyLayer(id: string) {
        if (!layers[id]) {
            layers[id] = {
                startMarker: null,
                endMarker: null,
                source: null,
                layer: null,
                listener: {
                    onClick: null,
                    onEnter: null,
                    onLeave: null,
                },
            };
        }
    }

    function addTrailLayer(
        trail: Trail,
        id: string,
        geojson: GeoJSON | null | undefined,
    ) {
        if (!geojson || !map) {
            return;
        }

        createEmptyLayer(id);

        if (!layers[id].source) {
            try {
                map.addSource(id, {
                    type: "geojson",
                    data: geojson,
                });
                layers[id].source = map.getSource(id) as M.GeoJSONSource;
                // map.addSource("trail-source", {
                //     type: "vector",
                //     url: `http://localhost:8080/data/out.json`,
                // });
            } catch (e) {
                return;
            }
        } else {
            layers[id].source.setData(geojson);
        }

        if (!layers[id].layer) {
            map.addLayer({
                id: id,
                type: "line",
                source: id,
                minzoom: minZoom,
                paint: {
                    "line-color": "#648ad5",
                    "line-width": 5,
                },
            });
            layers[id].layer = map.getLayer(id) as M.LineLayerSpecification;
            layers[id].listener.onEnter = (e) => highlightTrail(id);
            layers[id].listener.onLeave = (e) => unHighlightTrail(id);
            layers[id].listener.onClick = (e) => focusTrail(trail, e);

            map.on("mouseenter", id, layers[id].listener.onEnter);
            map.on("mouseleave", id, layers[id].listener.onLeave);
            map.on("click", id, layers[id].listener.onClick);

            // map.addLayer({
            //     id: "trail-layer",
            //     type: "line",
            //     source: "trail-source",
            //     "source-layer": "Herzogstand", // Replace with the actual layer name in the .mbtiles file
            //     layout: {
            //         "line-join": "round",
            //         "line-cap": "round",
            //     },
            //     paint: {
            //         "line-color": "#648ad5",
            //         "line-width": 5,
            //     },
            // });
        }

        if (!drawing) {
            addStartEndMarkers(trail, id, geojson);
        }
    }

    export function highlightTrail(id: string) {
        map?.setPaintProperty(id, "line-width", 7);
        map?.setPaintProperty(id, "line-color", "#2766e3");
    }

    export function unHighlightTrail(id: string) {
        map?.setPaintProperty(id, "line-width", 5);
        map?.setPaintProperty(id, "line-color", "#648ad5");
    }

    export function focusTrail(trail: Trail, e?: M.MapMouseEvent) {
        const currentlyFocussedTrail = trails[activeTrail];
        if (currentlyFocussedTrail) {
            unFocusTrail(currentlyFocussedTrail);
        }
        e?.preventDefault();
        dispatch("select", trail);
        const index = trails.findIndex((t) => t.id == trail.id);
        if (index == -1) {
            return;
        }
        activeTrail = index;

        highlightTrail(trail.id!);
        if (data[activeTrail] && showElevation) {
            epc?.setData(
                data[activeTrail]!,
                trails.at(activeTrail)!.expand.waypoints,
            );
            epc?.showProfile();
        }
        showWaypoints();
        flyToBounds();
    }

    export function unFocusTrail(trail: Trail) {
        dispatch("unselect", trail);
        activeTrail = -1;
        unHighlightTrail(trail.id!);
        flyToBounds();

        if (showElevation) {
            epc?.hideProfile();
        }
        hideWaypoints();
    }

    function addStartEndMarkers(
        trail: Trail,
        id: string,
        geojson: GeoJSON | null | undefined,
    ) {
        if (!map || !trail) {
            return;
        }
        createEmptyLayer(id);

        layers[id].startMarker ??= new FontawesomeMarker(
            { icon: "fa fa-bullseye" },
            {},
        );

        if (!geojson) {
            if (trail.lon && trail.lat) {
                layers[id].startMarker
                    .setLngLat([trail.lon, trail.lat])
                    .addTo(map);
            }
            return;
        }

        const startEndPoint = findStartAndEndPoints(geojson);

        layers[id].startMarker
            .setLngLat(startEndPoint[0] as M.LngLatLike)
            .addTo(map);

        if (showInfoPopup) {
            const popup = createPopupFromTrail(trail);
            layers[id].startMarker.setPopup(popup);
        }

        layers[id].endMarker ??= new FontawesomeMarker(
            { icon: "fa fa-flag-checkered" },
            {},
        );
        layers[id].endMarker.setLngLat(startEndPoint[1] as M.LngLatLike);
        if (map.getZoom() > minZoom) {
            layers[id].endMarker.addTo(map);
        }
    }

    export function togglePopup(id: string) {
        layers[id]?.startMarker?.togglePopup();
    }

    function showWaypoints() {
        if (!map) {
            return;
        }
        for (const waypoint of trails[activeTrail]?.expand.waypoints ?? []) {
            const marker = createMarkerFromWaypoint(waypoint);
            marker.addTo(map);
            markers.push(marker);
        }
    }

    function hideWaypoints() {
        if (!map) {
            return;
        }
        for (const m of markers) {
            m.remove()
        }
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

        const mapStyles: { text: string; value: string; thumbnail?: string }[] =
            [
                ...(($page.data.settings as Settings).tilesets ?? []).map(
                    (t) => ({
                        text: t.name,
                        value: t.url,
                    }),
                ),
                {
                    text: "Open Street Maps",
                    value: "/styles/osm.json",
                    thumbnail: "https://tile.openstreetmap.org/1/0/0.png",
                },
                {
                    text: "Open Topo Maps",
                    value: "/styles/otm.json",
                    thumbnail: "https://tile.opentopomap.org/1/0/0.png",
                },
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
        let preferredMapStyleIndex = mapStyles.findIndex(
            (s) => s.text === localStorage.getItem("layer"),
        );

        if (preferredMapStyleIndex == -1) {
            preferredMapStyleIndex = 0;
        }

        const finalMapOptions: M.MapOptions = {
            ...{
                container: mapContainer,
                style:
                    mapStyles[preferredMapStyleIndex].value ??
                    mapStyles[0].value,
                center: [initialState.lng, initialState.lat],
                zoom: initialState.zoom,
            },
            ...mapOptions,
        };
        map = new M.Map(finalMapOptions);

        const elevationMarker = new FontawesomeMarker(
            {
                id: "elevation-marker",
                icon: "fa-regular fa-circle",
                fontSize: "xs",
                width: 4,
                backgroundColor: "bg-primary",
                fontColor: "white",
            },
            {},
        );
        elevationMarker.setLngLat([0, 0]).addTo(map);
        elevationMarker.setOpacity("0");

        const switcherControl = new StyleSwitcherControl({
            styles: mapStyles,
            onSwitch: (style) => {
                layers = {};
                map?.setStyle(style.value);
                localStorage.setItem("layer", style.text);
            },
            selectedIndex:
                preferredMapStyleIndex !== -1 ? preferredMapStyleIndex : 0,
        });
        map.addControl(
            new M.NavigationControl({ visualizePitch: showTerrain }),
        );
        map.addControl(
            new M.ScaleControl({
                maxWidth: 120,
                unit: $page.data.settings?.unit ?? "metric",
            }),
            "top-left",
        );

        if (showElevation) {
            epc = new ElevationProfileControl({
                visible: false,
                profileBackgroundColor:
                    $theme == "light" ? "#242734" : "#191b24",
                backgroundColor: "bg-menu-background/90",
                unit: $page.data.settings?.unit ?? "metric",
                profileLineWidth: 3,
                displayDistanceGrid: true,
                tooltipDisplayDPlus: false,
                zoom: false,
                container: elevationProfileContainer,
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
            map.addControl(epc);
        }
        if (showStyleSwitcher) {
            map.addControl(switcherControl);
        }

        if (showFullscreen) {
            map.addControl(
                new FullscreenControl(() => {
                    dispatch("fullscreen");
                }),
                "bottom-right",
            );
        }

        if (showTerrain) {
            map!.addControl(
                new M.TerrainControl({
                    source: "terrain",
                }),
            );
        }

        map.on("styledata", () => {
            trails.forEach((t, i) => {
                addTrailLayer(t, t.id ?? i.toString(), data?.at(i));
            });

            if (showTerrain && $page.data.settings?.terrain && !map?.getSource("terrain")) {
                map!.addSource("terrain", {
                    type: "raster-dem",
                    url: $page.data.settings.terrain,
                });
            }
        });

        map.on("moveend", (e) => {
            dispatch("moveend", e.target);
        });

        map.on("zoom", (e) => {
            const zoom = e.target.getZoom();
            Object.values(layers).forEach((l) => {
                if (zoom > minZoom && map) {
                    l.endMarker?.addTo(map);
                } else {
                    l.endMarker?.remove();
                }
            });

            dispatch("zoom", e.target);
        });

        map.on("click", (e) => {
            dispatch("click", e);
        });

        map.on("load", () => {
            dispatch("init", map);
        });

        showWaypoints();
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

    :global(.maplibregl-popup-content) {
        @apply bg-background rounded-md;
    }
</style>
