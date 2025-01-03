<script lang="ts">
    import { page } from "$app/stores";
    import { PUBLIC_VALHALLA_URL } from "$env/static/public";
    import Button from "$lib/components/base/button.svelte";
    import Datepicker from "$lib/components/base/datepicker.svelte";
    import Select from "$lib/components/base/select.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import Textarea from "$lib/components/base/textarea.svelte";
    import Toggle from "$lib/components/base/toggle.svelte";
    import ListSelectModal from "$lib/components/list/list_select_modal.svelte";
    import SummitLogCard from "$lib/components/summit_log/summit_log_card.svelte";
    import SummitLogModal from "$lib/components/summit_log/summit_log_modal.svelte";
    import MapWithElevationMaplibre from "$lib/components/trail/map_with_elevation_maplibre.svelte";
    import PhotoPicker from "$lib/components/trail/photo_picker.svelte";
    import WaypointCard from "$lib/components/waypoint/waypoint_card.svelte";
    import WaypointModal from "$lib/components/waypoint/waypoint_modal.svelte";
    import { SummitLogCreateSchema } from "$lib/models/api/summit_log_schema.js";
    import { TrailCreateSchema } from "$lib/models/api/trail_schema.js";
    import { WaypointCreateSchema } from "$lib/models/api/waypoint_schema.js";
    import GPX from "$lib/models/gpx/gpx";
    import type { List } from "$lib/models/list";
    import { SummitLog } from "$lib/models/summit_log";
    import { Trail } from "$lib/models/trail";
    import type { ValhallaAnchor } from "$lib/models/valhalla";
    import { Waypoint } from "$lib/models/waypoint";
    import { categories } from "$lib/stores/category_store";
    import {
        lists_add_trail,
        lists_remove_trail,
    } from "$lib/stores/list_store";
    import { summitLog } from "$lib/stores/summit_log_store";
    import { show_toast } from "$lib/stores/toast_store";
    import {
        trail,
        trails_create,
        trails_update,
    } from "$lib/stores/trail_store.js";
    import {
        anchors,
        appendToRoute,
        calculateRouteBetween,
        clearRoute,
        deleteFromRoute,
        editRoute,
        route,
        setRoute,
    } from "$lib/stores/valhalla_store";
    import { waypoint } from "$lib/stores/waypoint_store";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { fromFile, gpx2trail } from "$lib/util/gpx_util";

    import {
        createAnchorMarker,
        createMarkerFromWaypoint,
    } from "$lib/util/maplibre_util";
    import { validator } from "@felte/validator-zod";
    import cryptoRandomString from "crypto-random-string";
    import { createForm } from "felte";
    import * as M from "maplibre-gl";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import { backInOut } from "svelte/easing";
    import { scale } from "svelte/transition";
    import { z } from "zod";

    export let data;

    let map: M.Map;
    let mapTrail: Trail[] = [];
    $: gpxData = $formData.expand!?.gpx_data;
    $: if (gpxData) {
        updateTrailOnMap();
    }

    let openWaypointModal: () => void;
    let openSummitLogModal: () => void;
    let openListSelectModal: () => void;

    let loading = false;

    let editingBasicInfo: boolean = false;

    let photoFiles: File[] = [];

    let gpxFile: File | Blob | null = null;

    let drawingActive = false;
    let overwriteGPX = false;

    const ClientTrailCreateSchema = TrailCreateSchema.extend({
        expand: z
            .object({
                gpx_data: z.string().optional(),
                summit_logs: z.array(SummitLogCreateSchema),
                waypoints: z.array(
                    WaypointCreateSchema.extend({
                        marker: z.any().optional(),
                    }),
                ),
            })
            .optional(),
    });

    const modesOfTransport = [
        { text: $_("hiking"), value: "pedestrian" },
        { text: $_("cycling"), value: "bicycle" },
        { text: $_("driving"), value: "auto" },
    ];
    let selectedModeOfTransport = modesOfTransport[0].value;

    let autoRouting = true;

    const {
        form,
        errors,
        data: formData,
        setFields,
    } = createForm<z.infer<typeof ClientTrailCreateSchema>>({
        initialValues: {
            ...data.trail,
            public: data.trail.id
                ? data.trail.public
                : $page.data.settings?.privacy.trails === "public",
            category:
                data.trail.category ||
                $page.data.settings?.category ||
                $categories[0].id,
        },
        extend: validator({
            schema: ClientTrailCreateSchema,
        }),
        onSubmit: async (form) => {
            loading = true;
            try {
                const htmlForm = document.getElementById(
                    "trail-form",
                ) as HTMLFormElement;
                const formData = new FormData(htmlForm);
                if (!formData.get("public")) {
                    form.public = false;
                }
                form.photos = form.photos.filter(
                    (p) => !p.startsWith("data:image/svg+xml;base64"),
                );

                if ($formData.expand!.gpx_data && overwriteGPX) {
                    gpxFile = new Blob([$formData.expand!.gpx_data], {
                        type: "text/xml",
                    });
                }

                if (
                    (!$formData.lat || !$formData.lon) &&
                    route.trk?.at(0)?.trkseg?.at(0)?.trkpt?.at(0)
                ) {
                    $formData.lat = route.trk
                        ?.at(0)
                        ?.trkseg?.at(0)
                        ?.trkpt?.at(0)?.$.lat;
                    $formData.lon = route.trk
                        ?.at(0)
                        ?.trkseg?.at(0)
                        ?.trkpt?.at(0)?.$.lon;
                }

                if (!form.id) {
                    const createdTrail = await trails_create(
                        form as Trail,
                        photoFiles,
                        gpxFile,
                    );
                    $formData.id = createdTrail.id;
                } else {
                    await trails_update(
                        $trail,
                        form as Trail,
                        photoFiles,
                        gpxFile,
                    );
                }

                show_toast({
                    type: "success",
                    icon: "check",
                    text: $_("trail-saved-successfully"),
                });
            } catch (e) {
                console.error(e);

                show_toast({
                    type: "error",
                    icon: "close",
                    text: $_("error-saving-trail"),
                });
            } finally {
                loading = false;
            }
        },
    });

    onMount(async () => {
        clearAnchorMarker();
        clearRoute();

        if ($formData.expand!.gpx_data) {
            const gpx = await GPX.parse($formData.expand!.gpx_data);
            if (!(gpx instanceof Error)) {
                setRoute(gpx);
                initRouteAnchors(gpx);
            }
        }
    });

    function openFileBrowser() {
        document.getElementById("fileInput")!.click();
    }

    async function handleFileSelection() {
        const selectedFile = (
            document.getElementById("fileInput") as HTMLInputElement
        ).files?.[0];

        if (!selectedFile) {
            return;
        }

        clearWaypoints();
        clearAnchorMarker();
        clearRoute();
        drawingActive = false;
        overwriteGPX = false;

        const { gpxData, gpxFile: file } = await fromFile(selectedFile);
        gpxFile = file;

        try {
            const prevId = $formData.id;
            const parseResult = await gpx2trail(gpxData, selectedFile.name);
            setFields(parseResult.trail);
            $formData.id = prevId;
            $formData.expand!.gpx_data = gpxData;
            setFields(
                "category",
                $page.data.settings.category || $categories[0].id,
            );
            setFields(
                "public",
                $page.data.settings?.privacy.trails === "public",
            );

            const log = new SummitLog(parseResult.trail.date as string, {
                distance: $formData.distance,
                elevation_gain: $formData.elevation_gain,
                elevation_loss: $formData.elevation_loss,
                duration: $formData.duration
                    ? $formData.duration * 60
                    : undefined,
            });

            log.expand!.gpx_data = gpxData;
            const blob = new Blob([gpxData], { type: selectedFile.type });
            log._gpx = new File([blob], selectedFile.name, {
                type: selectedFile.type,
            });

            $formData.expand!.summit_logs.push(log);

            setRoute(parseResult.gpx);
            initRouteAnchors(parseResult.gpx);

            for (const waypoint of $formData.expand!.waypoints) {
                saveWaypoint(waypoint);
            }
        } catch (e) {
            console.error(e);

            show_toast({
                icon: "close",
                type: "error",
                text: $_("error-reading-file"),
            });
            return;
        }
        const r = await fetch("/api/v1/search/cities500", {
            method: "POST",
            body: JSON.stringify({
                q: "",
                options: {
                    filter: [
                        `_geoRadius(${$formData.lat}, ${$formData.lon}, 10000)`,
                    ],
                    sort: [`_geoPoint(${$formData.lat}, ${$formData.lon}):asc`],
                    limit: 1,
                },
            }),
        });
        const closestCity = (await r.json()).hits[0];

        setFields("location", closestCity.name);
    }

    function clearWaypoints() {
        for (const waypoint of $formData.expand!.waypoints) {
            waypoint.marker?.remove();
        }
        $formData.expand!.waypoints = [];
        $formData.waypoints = [];
    }

    function clearAnchorMarker() {
        for (const anchor of anchors) {
            anchor.marker?.remove();
        }
    }

    function initRouteAnchors(gpx: GPX) {
        const segments = gpx.trk?.at(0)?.trkseg ?? [];

        for (let i = 0; i < segments.length; i++) {
            const segment = segments[i];
            const points = segment.trkpt ?? [];

            if (points.length > 0) {
                addAnchor(points[0].$.lat!, points[0].$.lon!, false);
            }
            if (i == segments.length - 1) {
                addAnchor(
                    points[points.length - 1].$.lat!,
                    points[points.length - 1].$.lon!,
                    false,
                );
            }
        }
    }

    function openMarkerPopup(waypoint: Waypoint) {
        waypoint.marker?.togglePopup();
    }

    function handleWaypointMenuClick(
        currentWaypoint: Waypoint,
        index: number,
        e: CustomEvent<{ text: string; value: string }>,
    ) {
        if (e.detail.value === "edit") {
            waypoint.set(currentWaypoint);
            openWaypointModal();
        } else if (e.detail.value === "delete") {
            currentWaypoint.marker?.remove();
            deleteWaypoint(index);
        }
    }

    function beforeWaypointModalOpen() {
        const mapCenter = map.getCenter();
        waypoint.set(new Waypoint(mapCenter.lat, mapCenter.lng));
        openWaypointModal();
    }

    function deleteWaypoint(index: number) {
        $formData.expand!.waypoints.splice(index, 1);
        $formData.waypoints.splice(index, 1);
        $formData.expand!.waypoints = $formData.expand!.waypoints;
    }

    function saveWaypoint(savedWaypoint: Waypoint) {
        let editedWaypointIndex = $formData.expand!.waypoints.findIndex(
            (s) => s.id == savedWaypoint.id,
        );

        if (editedWaypointIndex >= 0) {
            $formData.expand!.waypoints[editedWaypointIndex].marker?.remove();
            $formData.expand!.waypoints[editedWaypointIndex] = savedWaypoint;
        } else {
            savedWaypoint.id = cryptoRandomString({ length: 15 });
            $formData.expand!.waypoints = [
                ...$formData.expand!.waypoints,
                savedWaypoint,
            ];
        }
        const marker = createMarkerFromWaypoint(savedWaypoint, moveMarker);

        marker.addTo(map);
        savedWaypoint.marker = marker;
    }

    function moveMarker(marker: M.Marker, wpId?: string) {
        const position = marker.getLngLat();
        const editableWaypointIndex = $formData.expand!.waypoints.findIndex(
            (w) => w.id == wpId,
        );
        const editableWaypoint =
            $formData.expand!.waypoints[editableWaypointIndex];
        if (!editableWaypoint) {
            return;
        }
        editableWaypoint.lat = position.lat;
        editableWaypoint.lon = position.lng;
        $formData.expand!.waypoints = [...$formData.expand!.waypoints];
    }

    function beforeSummitLogModalOpen() {
        summitLog.set(new SummitLog(new Date().toISOString().split("T")[0]));
        openSummitLogModal();
    }

    function saveSummitLog(e: CustomEvent<SummitLog>) {
        const savedSummitLog = e.detail;

        let editedSummitLogIndex = $formData.expand!.summit_logs.findIndex(
            (s) => s.id == savedSummitLog.id,
        );

        if (editedSummitLogIndex >= 0) {
            $formData.expand!.summit_logs[editedSummitLogIndex] =
                savedSummitLog;
        } else {
            savedSummitLog.id = cryptoRandomString({ length: 15 });
            $formData.expand!.summit_logs = [
                ...$formData.expand!.summit_logs,
                savedSummitLog,
            ];
        }
    }

    function handleSummitLogMenuClick(
        currentSummitLog: SummitLog,
        index: number,
        e: CustomEvent<{ text: string; value: string }>,
    ) {
        if (e.detail.value === "edit") {
            summitLog.set(currentSummitLog);
            openSummitLogModal();
        } else if (e.detail.value === "delete") {
            $formData.expand!.summit_logs.splice(index, 1);
            $formData.summit_logs.splice(index, 1);
            $formData.expand!.summit_logs = $formData.expand!.summit_logs;
        }
    }

    async function handleListSelection(list: List) {
        if (!$formData.id) {
            return;
        }
        try {
            if (list.trails?.includes($formData.id!)) {
                list = await lists_remove_trail(list, $formData as Trail);
            } else {
                list = await lists_add_trail(list, $formData as Trail);
            }
            const index = data.lists.items.findIndex((l) => l.id == list.id);
            if (index >= 0) {
                data.lists.items[index] = list;
            }
            // await lists_index({ q: "", author: $currentUser?.id ?? "" }, 1, -1);
        } catch (e) {
            console.error(e);
            show_toast({
                type: "error",
                icon: "close",
                text: "Error adding trail to list.",
            });
        }
    }

    function startDrawing() {
        drawingActive = true;
        if (!route.trk?.at(0)?.trkseg?.at(0)?.trkpt?.length) {
        }
        for (const anchor of anchors) {
            anchor.marker?.addTo(map);
        }
    }

    function stopDrawing() {
        drawingActive = false;
        for (const anchor of anchors) {
            anchor.marker?.remove();
        }
    }

    async function handleMapClick(e: M.MapMouseEvent) {
        if (!drawingActive) {
            return;
        }
        const anchorCount = anchors.length;
        if (anchorCount == 0) {
            addAnchor(e.lngLat.lat, e.lngLat.lng);
        } else {
            const previousAnchor = anchors[anchorCount - 1];
            try {
                const routeWaypoints = await calculateRouteBetween(
                    previousAnchor.lat,
                    previousAnchor.lon,
                    e.lngLat.lat,
                    e.lngLat.lng,
                    selectedModeOfTransport,
                    autoRouting,
                );
                appendToRoute(routeWaypoints);
                addAnchor(e.lngLat.lat, e.lngLat.lng);
                updateTrailWithRouteData();
            } catch (e) {
                console.error(e);
                show_toast({
                    text: "Error calculating route",
                    icon: "close",
                    type: "error",
                });
            }
        }
    }

    function addAnchor(lat: number, lon: number, addtoMap: boolean = true) {
        const anchor: ValhallaAnchor = {
            id: cryptoRandomString({ length: 15 }),
            lat: lat,
            lon: lon,
        };
        const marker = createAnchorMarker(
            lat,
            lon,
            anchors.length + 1,
            () => {
                removeAnchor(anchors.findIndex((a) => a.id == anchor.id));
            },
            (_) => {
                if (!drawingActive) {
                    return;
                }
                const position = marker.getLngLat();
                anchor.lat = position.lat;
                anchor.lon = position.lng;
                recalculateRoute(anchors.findIndex((a) => a.id == anchor.id));
            },
        );
        if (addtoMap) {
            marker.addTo(map);
        }
        anchor.marker = marker;
        anchors.push(anchor);
    }

    function removeAnchor(anchorIndex: number) {
        if (!drawingActive) {
            return;
        }
        anchors[anchorIndex]?.marker?.remove();
        anchors.splice(anchorIndex, 1);
        for (let i = anchorIndex; i < anchors.length; i++) {
            const anchor = anchors[i];
            const markerIcon = anchor.marker?.getElement();
            if (markerIcon) {
                const markerText = markerIcon.textContent ?? "0";
                const markerIndex = parseInt(markerText);
                markerIcon.textContent = markerIndex - 1 + "";
            }
        }
        if (anchorIndex == 0) {
            deleteFromRoute(anchorIndex);
            updateTrailWithRouteData();
        } else if (anchorIndex == anchors.length) {
            deleteFromRoute(anchorIndex - 1);
            updateTrailWithRouteData();
        } else {
            deleteFromRoute(anchorIndex - 1);
            recalculateRoute(anchorIndex);
        }
    }

    async function recalculateRoute(anchorIndex: number) {
        const anchor = anchors[anchorIndex];
        if (!anchor) {
            return;
        }
        let nextRouteSegment;
        let previousRouteSegment;
        if (anchorIndex < anchors.length - 1) {
            const nextAnchor = anchors[anchorIndex + 1];

            nextRouteSegment = await calculateRouteBetween(
                anchor.lat,
                anchor.lon,
                nextAnchor.lat,
                nextAnchor.lon,
                selectedModeOfTransport,
                autoRouting,
            );
        }
        if (anchorIndex > 0) {
            const previousAnchor = anchors[anchorIndex - 1];
            previousRouteSegment = await calculateRouteBetween(
                previousAnchor.lat,
                previousAnchor.lon,
                anchor.lat,
                anchor.lon,
                selectedModeOfTransport,
                autoRouting,
            );
        }
        if (nextRouteSegment) {
            editRoute(anchorIndex, nextRouteSegment);
        }
        if (previousRouteSegment) {
            editRoute(anchorIndex - 1, previousRouteSegment);
        }
        updateTrailWithRouteData();
    }

    function updateTrailWithRouteData() {
        overwriteGPX = true;
        const totals = route.getTotals();
        $formData.distance = totals.distance;
        $formData.duration = totals.duration;
        $formData.elevation_gain = totals.elevationGain;
        $formData.elevation_loss = totals.elevationLoss;
        $formData.expand!.gpx_data = route.toString();
    }

    function updateTrailOnMap() {
        mapTrail = [{ ...($formData as Trail) }];
    }
</script>

<svelte:head>
    <title
        >{$formData.id ? `${$formData.name} | ${$_("edit")}` : $_("new-trail")} |
        wanderer</title
    >
</svelte:head>

<main class="grid grid-cols-1 md:grid-cols-[400px_1fr]">
    <form
        id="trail-form"
        class="overflow-y-auto overflow-x-hidden flex flex-col gap-4 px-8 order-1 md:order-none mt-8 md:mt-0"
        use:form
    >
        <h3 class="text-xl font-semibold">{$_("pick-a-trail")}</h3>
        <Button
            primary={true}
            type="button"
            disabled={drawingActive}
            on:click={openFileBrowser}>{$_("upload-file")}</Button
        >
        {#if PUBLIC_VALHALLA_URL}
            <div class="flex gap-4 items-center w-full">
                <hr class="basis-full border-input-border" />
                <span class="text-gray-500 uppercase">{$_("or")}</span>
                <hr class="basis-full border-input-border" />
            </div>
            <button
                class="btn-primary"
                type="button"
                on:click={drawingActive ? stopDrawing : startDrawing}
            >
                {drawingActive
                    ? $_("stop-drawing")
                    : $_("draw-a-route")}</button
            >
        {/if}
        <input
            type="file"
            name="gpx"
            id="fileInput"
            accept=".gpx,.GPX,.tcx,.TCX,.kml,.KML,.fit,.FIT"
            style="display: none;"
            on:change={handleFileSelection}
        />
        <hr class="border-separator" />
        <div class="flex gap-x-2">
            <h3 class="text-xl font-semibold">{$_("basic-info")}</h3>
            <button
                type="button"
                class="btn-icon"
                style="font-size: 0.9rem"
                on:click={() => (editingBasicInfo = !editingBasicInfo)}
                ><i class="fa fa-{editingBasicInfo ? 'check' : 'pen'}"
                ></i></button
            >
        </div>

        <fieldset
            class="grid grid-cols-2 gap-4 justify-around"
            data-felte-keep-on-remove
        >
            {#if editingBasicInfo}
                <TextField
                    bind:value={$formData.distance}
                    name="distance"
                    label={$_("distance")}
                ></TextField>
                <TextField
                    bind:value={$formData.duration}
                    name="duration"
                    label={$_("est-duration")}
                ></TextField><TextField
                    bind:value={$formData.elevation_gain}
                    name="elevation_gain"
                    label={$_("elevation-gain")}
                ></TextField>
                <TextField
                    bind:value={$formData.elevation_loss}
                    name="elevation_loss"
                    label={$_("elevation-loss")}
                ></TextField>
            {:else}
                <div>
                    <p>{$_("distance")}</p>
                    <span class="font-medium"
                        >{formatDistance($formData.distance)}</span
                    >
                    <input
                        type="hidden"
                        name="distance"
                        value={$formData.distance}
                    />
                </div>
                <div>
                    <p>{$_("est-duration")}</p>
                    <span class="font-medium"
                        >{formatTimeHHMM($formData.duration)}</span
                    >
                    <input
                        type="hidden"
                        name="duration"
                        value={$formData.duration}
                    />
                </div>
                <div>
                    <p>{$_("elevation-gain")}</p>
                    <span class="font-medium"
                        >{formatElevation($formData.elevation_gain)}</span
                    >
                    <input
                        type="hidden"
                        name="elevation_gain"
                        value={$formData.elevation_gain}
                    />
                </div>
                <div>
                    <p>{$_("elevation-loss")}</p>
                    <span class="font-medium"
                        >{formatElevation($formData.elevation_loss)}</span
                    >
                    <input
                        type="hidden"
                        name="elevation_gain"
                        value={$formData.elevation_gain}
                    />
                </div>
            {/if}
        </fieldset>
        <TextField name="name" label={$_("name")} error={$errors.name}
        ></TextField>
        <TextField
            name="location"
            label={$_("location")}
            error={$errors.location}
        ></TextField>
        <Datepicker label={$_("date")} bind:value={$formData.date}></Datepicker>
        <Textarea name="description" label={$_("describe-your-trail")}
        ></Textarea>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-y-4">
            <Select
                name="difficulty"
                label={$_("difficulty")}
                items={[
                    { text: $_("easy"), value: "easy" },
                    { text: $_("moderate"), value: "moderate" },
                    { text: $_("difficult"), value: "difficult" },
                ]}
            ></Select>
            <Select
                name="category"
                label={$_("category")}
                items={$categories.map((c) => ({
                    text: $_(c.name),
                    value: c.id,
                }))}
            ></Select>
        </div>

        <Toggle name="public" label={$_("public")}></Toggle>
        <hr class="border-separator" />
        <h3 class="text-xl font-semibold">
            {$_("waypoints", { values: { n: 2 } })}
        </h3>
        <ul>
            {#each $formData.expand?.waypoints ?? [] as waypoint, i}
                <li on:mouseenter={() => openMarkerPopup(waypoint)}>
                    <WaypointCard
                        {waypoint}
                        mode="edit"
                        on:change={(e) =>
                            handleWaypointMenuClick(waypoint, i, e)}
                    ></WaypointCard>
                </li>
            {/each}
        </ul>
        <button
            class="btn-secondary"
            type="button"
            on:click={beforeWaypointModalOpen}
            ><i class="fa fa-plus mr-2"></i>{$_("add-waypoint")}</button
        >
        <hr class="border-separator" />
        <h3 class="text-xl font-semibold">{$_("photos")}</h3>
        <PhotoPicker
            id="trail"
            parent={$formData}
            bind:photos={$formData.photos}
            bind:thumbnail={$formData.thumbnail}
            bind:photoFiles
        ></PhotoPicker>
        <hr class="border-separator" />
        <h3 class="text-xl font-semibold">{$_("summit-book")}</h3>
        <ul>
            {#each $formData.expand?.summit_logs ?? [] as log, i}
                <li>
                    <SummitLogCard
                        {log}
                        mode="edit"
                        on:change={(e) => handleSummitLogMenuClick(log, i, e)}
                    ></SummitLogCard>
                </li>
            {/each}
        </ul>
        <button
            class="btn-secondary"
            type="button"
            on:click={beforeSummitLogModalOpen}
            ><i class="fa fa-plus mr-2"></i>{$_("add-entry")}</button
        >
        {#if data.lists.items.length}
            <hr class="border-separator" />
            <h3 class="text-xl font-semibold">
                {$_("list", { values: { n: 2 } })}
            </h3>
            <div class="flex gap-4 flex-wrap">
                {#each data.lists.items as list}
                    {#if $formData.id && list.trails?.includes($formData.id)}
                        <div
                            class="flex gap-2 items-center border border-input-border rounded-xl p-2"
                        >
                            <img
                                class="w-8 aspect-square rounded-full object-cover"
                                src={list.avatar
                                    ? getFileURL(list, list.avatar)
                                    : "/imgs/default_list_thumbnail.webp"}
                                alt="avatar"
                            />

                            <span class="text-sm">{list.name}</span>
                        </div>
                    {/if}
                {/each}
            </div>
            <Button
                secondary={true}
                tooltip={$_("save-your-trail-first")}
                disabled={!$formData.id}
                type="button"
                on:click={openListSelectModal}
                ><i class="fa fa-plus mr-2"></i>{$_("add-to-list")}</Button
            >
        {/if}
        <hr class="border-separator" />
        <Button
            primary={true}
            large={true}
            type="submit"
            extraClasses="mb-2"
            {loading}>{$_("save-trail")}</Button
        >
    </form>
    <div class="relative">
        {#if drawingActive}
            <div
                class="absolute top-0 left-16 z-50 p-4 my-2 rounded-xl bg-background space-y-4"
                in:scale={{ easing: backInOut }}
                out:scale={{ easing: backInOut }}
            >
                <Toggle bind:value={autoRouting} label="Enable auto-routing"
                ></Toggle>
                <Select
                    items={modesOfTransport}
                    bind:value={selectedModeOfTransport}
                    disabled={!autoRouting}
                ></Select>
            </div>
        {/if}
        <MapWithElevationMaplibre
            trails={mapTrail}
            drawing={drawingActive}
            showTerrain={true}
            onMarkerDragEnd={moveMarker}
            bind:map
            on:click={(e) => handleMapClick(e.detail)}
        ></MapWithElevationMaplibre>
    </div>
</main>
<WaypointModal
    bind:openModal={openWaypointModal}
    on:save={(e) => saveWaypoint(e.detail)}
></WaypointModal>
<SummitLogModal bind:openModal={openSummitLogModal} on:save={saveSummitLog}
></SummitLogModal>
<ListSelectModal
    lists={data.lists.items}
    bind:openModal={openListSelectModal}
    on:change={(e) => handleListSelection(e.detail)}
></ListSelectModal>

<style>
    #map {
        height: calc(400px);
    }
    @media only screen and (min-width: 768px) {
        #map,
        form {
            height: calc(100vh - 124px);
        }
    }
</style>
