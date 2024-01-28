<script lang="ts">
    import Modal from "$lib/components/base/modal.svelte";
    import Select from "$lib/components/base/select.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import Textarea from "$lib/components/base/textarea.svelte";
    import WaypointCard from "$lib/components/waypoint/waypoint_card.svelte";
    import WaypointModal from "$lib/components/waypoint/waypoint_modal.svelte";
    import { Trail } from "$lib/models/trail";
    import { Waypoint } from "$lib/models/waypoint";
    import { categories } from "$lib/stores/category_store";
    import { trail } from "$lib/stores/trail_store";
    import { waypoint } from "$lib/stores/waypoint_store";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import type { GPX, Icon, LeafletEvent, Map } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import { onMount } from "svelte";
    import type { W } from "vitest/dist/reporters-rzC174PQ.js";

    let L: any;
    let map: Map;

    let gpxLayer: GPX;

    let openWayPointModal: () => void;

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");

        map = L.map("map").setView([0, 0], 2);

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);
    });

    function openFileBrowser() {
        document.getElementById("fileInput")!.click();
    }

    function handleFileSelection() {
        const selectedFile = (
            document.getElementById("fileInput") as HTMLInputElement
        ).files?.[0];

        if (!selectedFile) {
            return;
        }

        $trail.gpx = selectedFile?.name;

        var reader = new FileReader();

        reader.readAsText(selectedFile);

        reader.onload = function (e) {
            trail.set(new Trail(""));
            gpxLayer?.remove();
            gpxLayer = new L.GPX(e.target?.result, {
                async: true,
                marker_options: {
                    wptIcons: {
                        "": L.AwesomeMarkers.icon({
                            icon: "circle",
                            prefix: "fa",
                            markerColor: "cadetblue",
                            iconColor: "white",
                        }) as Icon,
                    },
                    startIcon: L.AwesomeMarkers.icon({
                        icon: "circle-half-stroke",
                        prefix: "fa",
                        markerColor: "cadetblue",
                        iconColor: "white",
                    }) as Icon,
                    endIcon: L.AwesomeMarkers.icon({
                        icon: "flag-checkered",
                        prefix: "fa",
                        markerColor: "cadetblue",
                        iconColor: "white",
                    }) as Icon,
                    startIconUrl: "",
                    endIconUrl: "",
                    shadowUrl: "",
                },
            })
                .on("addpoint", function (e: any) {
                    if (e.point_type === "start") {
                        e.point.setZIndexOffset(1000);
                    } else if (e.point_type === "waypoint") {
                        const waypoint = new Waypoint(
                            e.point._latlng.lat,
                            e.point._latlng.lng,
                            { name: e.point.options.title, marker: e.point },
                        );
                        $trail.expand.waypoints.push(waypoint);
                    }
                })
                .on("loaded", function (e: LeafletEvent) {
                    map.fitBounds(e.target.getBounds());
                    $trail.distance = e.target.get_distance();
                    $trail.elevation_gain = e.target.get_elevation_gain();
                    $trail.duration = e.target.get_total_time() / 1000 / 60;
                })
                .addTo(map);
        };
    }

    function openMarkerPopup(waypoint: Waypoint) {
        waypoint.marker?.openPopup();
    }

    function handleWaypointMenuClick(
        currentWaypoint: Waypoint,
        e: CustomEvent<{ text: string; value: string }>,
    ) {
        if (e.detail.value === "edit") {
            waypoint.set(currentWaypoint);
            openWayPointModal();
        } else if (e.detail.value === "delete") {
            currentWaypoint.marker?.remove();
            $trail.expand.waypoints = $trail.expand.waypoints.filter(
                (w) =>
                    w.lat !== currentWaypoint.lat &&
                    w.lon !== currentWaypoint.lon,
            );
        }
    }

    function findMapCenterAndOpenModal() {
        const mapCenter = map.getCenter();
        waypoint.set(new Waypoint(mapCenter.lat, mapCenter.lng));
        openWayPointModal();
    }

    function saveWaypoint(e: CustomEvent<Waypoint>) {
        const savedWaypoint = e.detail;

        if ($trail.expand.waypoints.includes(savedWaypoint)) {
            $trail.expand.waypoints = $trail.expand.waypoints;
            savedWaypoint.marker?.remove();
        } else {
            $trail.expand.waypoints = [
                ...$trail.expand.waypoints,
                savedWaypoint,
            ];
        }
        const marker = createMarkerFromWaypoint(L, savedWaypoint);
        marker.addTo(map);
        savedWaypoint.marker = marker;
    }
</script>

<main class="grid grid-cols-[400px_1fr]">
    <form class="overflow-scroll flex flex-col gap-4 px-8">
        <h3 class="text-xl font-semibold">Pick a trail</h3>
        <button class="btn-primary" on:click={openFileBrowser}
            >Upload GPX</button
        >
        <input
            type="file"
            id="fileInput"
            style="display: none;"
            on:change={handleFileSelection}
        />
        <hr />
        <h3 class="text-xl font-semibold">Basic Info</h3>
        <div class="flex gap-4 justify-around">
            <div class="flex flex-col items-center">
                <span>Distance</span>
                <span class="font-medium">{formatMeters($trail.distance)}</span>
            </div>
            <div class="flex flex-col items-center">
                <span>Elevation gain</span>
                <span class="font-medium"
                    >{formatMeters($trail.elevation_gain)}</span
                >
            </div>
            <div class="flex flex-col items-center">
                <span>Est. duration</span>
                <span class="font-medium"
                    >{formatTimeHHMM($trail.duration)}</span
                >
            </div>
        </div>
        <TextField label="Name*" bind:value={$trail.name}></TextField>
        <Textarea label="Describe your trail" bind:value={$trail.description}
        ></Textarea>
        <Select
            label="Category"
            bind:value={$trail.expand.category}
            items={$categories.map((c) => ({ text: c.name, value: c }))}
        ></Select>
        <hr />
        <h3 class="text-xl font-semibold">Waypoints</h3>
        <ul>
            {#each $trail.expand.waypoints as waypoint, i}
                <li on:mouseenter={() => openMarkerPopup(waypoint)}>
                    <WaypointCard
                        {waypoint}
                        mode="edit"
                        on:change={(e) => handleWaypointMenuClick(waypoint, e)}
                    ></WaypointCard>
                </li>
            {/each}
        </ul>
        <button class="btn-secondary" on:click={findMapCenterAndOpenModal}
            ><i class="fa fa-plus mr-2"></i>Add Waypoint</button
        >
    </form>
    <div class="rounded-xl" id="map"></div>
</main>
<WaypointModal bind:openModal={openWayPointModal} on:save={saveWaypoint}
></WaypointModal>

<style>
    #map,
    form {
        height: calc(100vh - 124px);
    }
</style>
