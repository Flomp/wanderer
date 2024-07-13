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
    import MapWithElevation from "$lib/components/trail/map_with_elevation.svelte";
    import PhotoPicker from "$lib/components/trail/photo_picker.svelte";
    import WaypointCard from "$lib/components/waypoint/waypoint_card.svelte";
    import WaypointModal from "$lib/components/waypoint/waypoint_modal.svelte";
    import GPX from "$lib/models/gpx/gpx";
    import type { List } from "$lib/models/list";
    import { SummitLog } from "$lib/models/summit_log";
    import { Trail } from "$lib/models/trail";
    import type { ValhallaAnchor } from "$lib/models/valhalla";
    import { Waypoint } from "$lib/models/waypoint";
    import { categories } from "$lib/stores/category_store";
    import {
        lists,
        lists_add_trail,
        lists_index,
        lists_remove_trail,
    } from "$lib/stores/list_store";
    import { summitLog } from "$lib/stores/summit_log_store";
    import { show_toast } from "$lib/stores/toast_store";
    import {
        trail,
        trails_create,
        trails_update,
    } from "$lib/stores/trail_store";
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
    import { fromFIT, fromKML, fromTCX, gpx2trail, isFITFile } from "$lib/util/gpx_util";
    import {
        createAnchorMarker,
        createMarkerFromWaypoint,
    } from "$lib/util/leaflet_util";
    import { createForm } from "$lib/vendor/svelte-form-lib";
    import cryptoRandomString from "crypto-random-string";
    import type { DivIcon, LeafletMouseEvent, Map } from "leaflet";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import { backInOut } from "svelte/easing";
    import { scale } from "svelte/transition";
    import { array, number, object, string } from "yup";

    export let data: { trail: Trail };

    let L: any;
    let map: Map;

    let openWaypointModal: () => void;
    let openSummitLogModal: () => void;
    let openListSelectModal: () => void;

    let loading = false;

    let editingBasicInfo: boolean = false;

    let photoFiles: File[] = [];

    let gpxFile: File | Blob | null = null;

    let drawingActive = false;
    let overwriteGPX = false;

    const trailSchema = object<Trail>({
        id: string().optional(),
        name: string().required($_("required")),
        location: string().optional(),
        distance: number().optional(),
        difficulty: string()
            .oneOf(["easy", "moderate", "difficult"])
            .optional(),
        elevation_gain: number().optional(),
        duration: number().optional(),
        thumbnail: string().optional(),
        photos: array(string()).optional(),
        gpx: string().optional(),
        description: string().optional(),
    });

    const modesOfTransport = [
        { text: $_("hiking"), value: "pedestrian" },
        { text: $_("cycling"), value: "bicycle" },
        { text: $_("driving"), value: "auto" },
    ];
    let selectedModeOfTransport = modesOfTransport[0].value;

    let autoRouting = true;

    const { form, errors, handleChange, handleSubmit } = createForm<Trail>({
        initialValues: {
            ...data.trail,
            category: $page.data.settings?.category || $categories[0].id,
        },
        validationSchema: trailSchema,
        onError: (errors) => {
            if (errors.name) {
                const nameInput = document.querySelector(
                    "input[name=name]",
                ) as HTMLElement;

                if (window.innerWidth < 768) {
                    window?.scroll({
                        top: nameInput.offsetTop - 24,
                        behavior: "smooth",
                    });
                } else {
                    const form = document.getElementById("trail-form");
                    form?.scroll({
                        top: nameInput.offsetTop - 130,
                        behavior: "smooth",
                    });
                }
            }
        },
        onSubmit: async (submittedTrail) => {
            loading = true;
            try {
                const htmlForm = document.getElementById(
                    "trail-form",
                ) as HTMLFormElement;
                const formData = new FormData(htmlForm);
                if (!formData.get("public")) {
                    submittedTrail.public = false;
                }
                submittedTrail.photos = submittedTrail.photos.filter(
                    (p) => !p.startsWith("data:image/svg+xml;base64"),
                );

                if ($form.expand.gpx_data && overwriteGPX) {
                    gpxFile = new Blob([$form.expand.gpx_data], {
                        type: "text/xml",
                    });
                }

                if (
                    (!$form.lat || !$form.lon) &&
                    route.trk?.at(0)?.trkseg?.at(0)?.trkpt?.at(0)
                ) {
                    $form.lat = route.trk
                        ?.at(0)
                        ?.trkseg?.at(0)
                        ?.trkpt?.at(0)?.$.lat;
                    $form.lon = route.trk
                        ?.at(0)
                        ?.trkseg?.at(0)
                        ?.trkpt?.at(0)?.$.lon;
                }

                if (!submittedTrail.id) {
                    const createdTrail = await trails_create(
                        submittedTrail,
                        photoFiles,
                        gpxFile,
                    );
                    $form.id = createdTrail.id;
                } else {
                    await trails_update(
                        $trail,
                        submittedTrail,
                        photoFiles,
                        gpxFile,
                    );
                }

                show_toast({
                    type: "success",
                    icon: "check",
                    text: $_('trail-saved-successfully'),
                });
            } catch (e) {
                console.error(e);

                show_toast({
                    type: "error",
                    icon: "close",
                    text: $_('error-saving-trail'),
                });
            } finally {
                loading = false;
            }
        },
    });

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet.awesome-markers");

        clearAnchorMarker();
        clearRoute();

        if ($form.expand.gpx_data) {
            const gpx = await GPX.parse($form.expand.gpx_data);
            if (!(gpx instanceof Error)) {
                setRoute(gpx);
                initRouteAnchors(gpx);
            }
        }
    });

    function openFileBrowser() {
        document.getElementById("fileInput")!.click();
    }

    async function parseFile(file: File) {
        const fileExtension = file.name.split(".").pop()?.toLowerCase();

        let gpxData = "";
        const fileContent = await file.text();
        const fileBuffer = await file.arrayBuffer();

        if (!isFITFile(fileBuffer)) {
            if (fileContent.includes("http://www.opengis.net/kml")) {
                gpxData = fromKML(fileContent);
                gpxFile = new Blob([gpxData], {
                    type: "application/gpx+xml",
                });
                return gpxData;
            } else if (fileContent.includes("TrainingCenterDatabase")) {
                gpxData = fromTCX(fileContent);
                gpxFile = new Blob([gpxData], {
                    type: "application/gpx+xml",
                });
            } else {
                gpxData = fileContent;
                gpxFile = file;
            }
        } else {
            gpxData = await fromFIT(fileBuffer);
            gpxFile = new Blob([gpxData], {
                type: "application/gpx+xml",
            });
        }

        return gpxData;
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

        let gpxData = await parseFile(selectedFile);

        try {
            const parseResult = await gpx2trail(gpxData);
            $form = parseResult.trail;
            $form.expand.gpx_data = gpxData;
            $form.category = $page.data.settings.category || $categories[0].id;

            setRoute(parseResult.gpx);
            initRouteAnchors(parseResult.gpx);

            for (const waypoint of $form.expand.waypoints) {
                saveWaypoint(waypoint);
            }
        } catch (e) {
            console.log(e);

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
                    filter: [`_geoRadius(${$form.lat}, ${$form.lon}, 10000)`],
                    sort: [`_geoPoint(${$form.lat}, ${$form.lon}):asc`],
                    limit: 1,
                },
            }),
        });
        const closestCity = (await r.json()).hits[0];

        $form.location = closestCity.name;
    }

    function clearWaypoints() {
        for (const waypoint of $form.expand.waypoints) {
            waypoint.marker?.remove();
        }
        $form.expand.waypoints = [];
        $form.waypoints = [];
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
        waypoint.marker?.openPopup();
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
        $form.expand.waypoints.splice(index, 1);
        $form.waypoints.splice(index, 1);
        $form.expand.waypoints = $form.expand.waypoints;
    }

    function saveWaypoint(savedWaypoint: Waypoint) {
        let editedWaypointIndex = $form.expand.waypoints.findIndex(
            (s) => s.id == savedWaypoint.id,
        );

        if (editedWaypointIndex >= 0) {
            $form.expand.waypoints[editedWaypointIndex].marker?.remove();
            $form.expand.waypoints[editedWaypointIndex] = savedWaypoint;
        } else {
            savedWaypoint.id = cryptoRandomString({ length: 15 });
            $form.expand.waypoints = [...$form.expand.waypoints, savedWaypoint];
        }
        const marker = createMarkerFromWaypoint(L, savedWaypoint, (event) => {
            var marker = event.target;
            var position = marker.getLatLng();
            const editableWaypointIndex = $form.expand.waypoints.findIndex(
                (w) => w.id == savedWaypoint.id,
            );
            const editableWaypoint =
                $form.expand.waypoints[editableWaypointIndex];
            editableWaypoint!.lat = position.lat;
            editableWaypoint!.lon = position.lng;
            $form.expand.waypoints = [...$form.expand.waypoints];
        });

        marker.addTo(map);
        savedWaypoint.marker = marker;
    }

    function beforeSummitLogModalOpen() {
        summitLog.set(new SummitLog(new Date().toISOString().split("T")[0]));
        openSummitLogModal();
    }

    function saveSummitLog(e: CustomEvent<SummitLog>) {
        const savedSummitLog = e.detail;
        let editedSummitLogIndex = $form.expand.summit_logs.findIndex(
            (s) => s.id == savedSummitLog.id,
        );

        if (editedSummitLogIndex >= 0) {
            $form.expand.summit_logs[editedSummitLogIndex] = savedSummitLog;
        } else {
            savedSummitLog.id = cryptoRandomString({ length: 15 });
            $form.expand.summit_logs = [
                ...$form.expand.summit_logs,
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
            $form.expand.summit_logs.splice(index, 1);
            $form.summit_logs.splice(index, 1);
            $form.expand.summit_logs = $form.expand.summit_logs;
        }
    }

    async function handleListSelection(list: List) {
        if (!$form.id) {
            return;
        }
        try {
            if (list.trails?.includes($form.id!)) {
                await lists_remove_trail(list, $form);
            } else {
                await lists_add_trail(list, $form);
            }
            await lists_index();
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

    async function handleMapClick(e: LeafletMouseEvent) {
        if (!drawingActive) {
            return;
        }
        const anchorCount = anchors.length;
        if (anchorCount == 0) {
            addAnchor(e.latlng.lat, e.latlng.lng);
        } else {
            const previousAnchor = anchors[anchorCount - 1];
            try {
                const routeWaypoints = await calculateRouteBetween(
                    previousAnchor.lat,
                    previousAnchor.lon,
                    e.latlng.lat,
                    e.latlng.lng,
                    selectedModeOfTransport,
                    autoRouting,
                );
                appendToRoute(routeWaypoints);
                addAnchor(e.latlng.lat, e.latlng.lng);
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
            L,
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
                const position = marker.getLatLng();
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
            const markerIcon = anchor.marker?.getIcon() as DivIcon | undefined;
            if (markerIcon) {
                const markerText =
                    (markerIcon.options.html as HTMLSpanElement).textContent ??
                    "0";
                const markerIndex = parseInt(markerText);
                (markerIcon.options.html as HTMLSpanElement).textContent =
                    markerIndex - 1 + "";
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
        $form.distance = totals.distance;
        $form.duration = totals.duration;
        $form.elevation_gain = totals.elevationGain;
        $form.expand.gpx_data = route.toString();
    }
</script>

<svelte:head>
    <title
        >{$form.id ? `${$form.name} | ${$_("edit")}` : $_("new-trail")} | wanderer</title
    >
</svelte:head>

<main class="grid grid-cols-1 md:grid-cols-[400px_1fr]">
    <form
        id="trail-form"
        class="overflow-y-auto overflow-x-hidden flex flex-col gap-4 px-8 order-1 md:order-none mt-8 md:mt-0"
        on:submit={handleSubmit}
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

        <div class="flex gap-4 justify-around">
            {#if editingBasicInfo}
                <TextField
                    bind:value={$form.distance}
                    name="distance"
                    label={$_("distance")}
                ></TextField>
                <TextField
                    bind:value={$form.elevation_gain}
                    name="elevation_gain"
                    label={$_("elevation-gain")}
                ></TextField>
                <TextField
                    bind:value={$form.duration}
                    name="duration"
                    label={$_("est-duration")}
                ></TextField>
            {:else}
                <div class="flex flex-col">
                    <span>{$_("distance")}</span>
                    <span class="font-medium"
                        >{formatDistance($form.distance)}</span
                    >
                    <input
                        type="hidden"
                        name="distance"
                        value={$form.distance}
                    />
                </div>
                <div class="flex flex-col">
                    <span>{$_("elevation-gain")}</span>
                    <span class="font-medium"
                        >{formatElevation($form.elevation_gain)}</span
                    >
                    <input
                        type="hidden"
                        name="elevation_gain"
                        value={$form.elevation_gain}
                    />
                </div>
                <div class="flex flex-col">
                    <span>{$_("est-duration")}</span>
                    <span class="font-medium"
                        >{formatTimeHHMM($form.duration)}</span
                    >
                    <input
                        type="hidden"
                        name="duration"
                        value={$form.duration}
                    />
                </div>
            {/if}
            <input type="hidden" name="lat" value={$form.lat} />
            <input type="hidden" name="lon" value={$form.lon} />
        </div>
        <TextField
            name="name"
            label={$_("name")}
            on:change={handleChange}
            error={$errors.name}
            bind:value={$form.name}
        ></TextField>
        <TextField
            name="location"
            label={$_("location")}
            error={$errors.location}
            bind:value={$form.location}
        ></TextField>
        <Datepicker label={$_("date")} bind:value={$form.date}></Datepicker>
        <Textarea
            name="description"
            label={$_("describe-your-trail")}
            bind:value={$form.description}
        ></Textarea>
        <Select
            name="difficulty"
            label={$_("difficulty")}
            bind:value={$form.difficulty}
            items={[
                { text: $_("easy"), value: "easy" },
                { text: $_("moderate"), value: "moderate" },
                { text: $_("difficult"), value: "difficult" },
            ]}
        ></Select>
        <Select
            name="category"
            label={$_("category")}
            bind:value={$form.category}
            items={$categories.map((c) => ({ text: $_(c.name), value: c.id }))}
        ></Select>
        <Toggle name="public" label={$_("public")} bind:value={$form.public}
        ></Toggle>
        <hr class="border-separator" />
        <h3 class="text-xl font-semibold">
            {$_("waypoints", { values: { n: 2 } })}
        </h3>
        <ul>
            {#each $form.expand.waypoints ?? [] as waypoint, i}
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
            parent={$form}
            bind:photos={$form.photos}
            bind:thumbnail={$form.thumbnail}
            bind:photoFiles
        ></PhotoPicker>
        <hr class="border-separator" />
        <h3 class="text-xl font-semibold">{$_("summit-book")}</h3>
        <ul>
            {#each $form.expand.summit_logs ?? [] as log, i}
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
        {#if $lists.length}
            <hr class="border-separator" />
            <h3 class="text-xl font-semibold">
                {$_("list", { values: { n: 2 } })}
            </h3>
            <div class="flex gap-4 flex-wrap">
                {#each $lists as list}
                    {#if $form.id && list.trails?.includes($form.id)}
                        <div
                            class="flex gap-2 items-center border border-input-border rounded-xl p-2"
                        >
                            {#if list.avatar}
                                <img
                                    class="w-8 aspect-square rounded-full object-cover"
                                    src={getFileURL(list, list.avatar)}
                                    alt="avatar"
                                />
                            {:else}
                                <i class="fa fa-table-list text-2xl"></i>
                            {/if}
                            <span class="text-sm">{list.name}</span>
                        </div>
                    {/if}
                {/each}
            </div>
            <Button
                secondary={true}
                tooltip={$_("save-your-trail-first")}
                disabled={!$form.id}
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
        <MapWithElevation
            trail={$form}
            crosshair={drawingActive}
            options={{
                autofitBounds: !drawingActive,
                mapTooltip: !drawingActive,
                speed: !drawingActive,
                speedFactor: drawingActive ? 0 : 1,
                showStartEnd: !drawingActive,
            }}
            bind:map
            on:click={(e) => handleMapClick(e.detail)}
        ></MapWithElevation>
    </div>
</main>
<WaypointModal
    bind:openModal={openWaypointModal}
    on:save={(e) => saveWaypoint(e.detail)}
></WaypointModal>
<SummitLogModal bind:openModal={openSummitLogModal} on:save={saveSummitLog}
></SummitLogModal>
<ListSelectModal
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
