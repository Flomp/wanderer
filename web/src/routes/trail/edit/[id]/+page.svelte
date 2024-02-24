<script lang="ts">
    import Button from "$lib/components/base/button.svelte";
    import Select from "$lib/components/base/select.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import Textarea from "$lib/components/base/textarea.svelte";
    import Toggle from "$lib/components/base/toggle.svelte";
    import PhotoCard from "$lib/components/photo_card.svelte";
    import SummitLogCard from "$lib/components/summit_log/summit_log_card.svelte";
    import SummitLogModal from "$lib/components/summit_log/summit_log_modal.svelte";
    import WaypointCard from "$lib/components/waypoint/waypoint_card.svelte";
    import WaypointModal from "$lib/components/waypoint/waypoint_modal.svelte";
    import { ms } from "$lib/meilisearch";
    import { SummitLog } from "$lib/models/summit_log";
    import { Trail, trailSchema } from "$lib/models/trail";
    import { Waypoint } from "$lib/models/waypoint";
    import { categories } from "$lib/stores/category_store";
    import { summitLog } from "$lib/stores/summit_log_store";
    import { show_toast } from "$lib/stores/toast_store";
    import {
        trail,
        trails_create,
        trails_update,
    } from "$lib/stores/trail_store";
    import { waypoint } from "$lib/stores/waypoint_store";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import { createForm } from "$lib/vendor/svelte-form-lib";
    import cryptoRandomString from "crypto-random-string";
    import { format } from "date-fns";
    import type { GPX, Icon, LeafletEvent, Map } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import { onMount } from "svelte";

    export let data: { trail: Trail };

    let L: any;
    let map: Map;

    let gpxLayer: GPX;

    let openWaypointModal: () => void;
    let openSummitLogModal: () => void;

    let loading = false;

    onMount(async () => {
        L = (await import("leaflet")).default;
        await import("leaflet-gpx");
        await import("leaflet.awesome-markers");

        map = L.map("map").setView([0, 0], 2);

        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors",
        }).addTo(map);

        if (
            data.trail.expand.gpx_data &&
            data.trail.expand.gpx_data.length > 0
        ) {
            addGPXLayer(data.trail.expand.gpx_data, false);
        }

        if (data.trail.expand.waypoints?.length > 0) {
            for (const waypoint of data.trail.expand.waypoints) {
                const marker = createMarkerFromWaypoint(L, waypoint);
                marker.addTo(map);
            }
        }
    });

    const { form, errors, handleChange, handleSubmit } = createForm<Trail>({
        initialValues: data.trail,
        validationSchema: trailSchema,
        onSubmit: async (submittedTrail) => {
            loading = true;
            try {
                const htmlForm = document.getElementById(
                    "trail-form",
                ) as HTMLFormElement;
                const formData = new FormData(htmlForm);
                if (!submittedTrail.id) {
                    const createdTrail = await trails_create(
                        submittedTrail,
                        formData,
                    );
                    $form.id = createdTrail.id;
                } else {
                    await trails_update($trail, submittedTrail, formData);
                }

                show_toast({
                    type: "success",
                    icon: "check",
                    text: "Trail saved successfully.",
                });
            } catch (e) {
                console.error(e);

                show_toast({
                    type: "error",
                    icon: "close",
                    text: "Error saving trail.",
                });
            } finally {
                loading = false;
            }
        },
    });

    function addGPXLayer(gpx: string, addWaypoints: boolean = true) {
        return new Promise<void>(function (resolve, reject) {
            gpxLayer?.remove();
            gpxLayer = new L.GPX(gpx, {
                async: true,
                gpx_options: {
                    parseElements: [
                        "track",
                        "route",
                        ...(addWaypoints ? ["waypoint"] : []),
                    ],
                },
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
                        $form.lat = e.point._latlng.lat;
                        $form.lon = e.point._latlng.lng;
                    } else if (e.point_type === "waypoint") {
                        const waypoint = new Waypoint(
                            e.point._latlng.lat,
                            e.point._latlng.lng,
                            { name: e.point.options.title, marker: e.point },
                        );
                        if (!$form.expand.waypoints) {
                        }
                        $form.expand.waypoints.push(waypoint);
                    }
                })
                .on("loaded", function (e: LeafletEvent) {
                    map.fitBounds(e.target.getBounds());
                    $form.distance = e.target.get_distance();
                    $form.elevation_gain = e.target.get_elevation_gain();
                    $form.duration = e.target.get_total_time() / 1000 / 60;
                    resolve();
                })
                .on("error", reject)
                .addTo(map);
        });
    }

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

        $form.gpx = selectedFile?.name;
        $form.expand.waypoints = [];

        var reader = new FileReader();

        reader.readAsText(selectedFile);

        reader.onload = async function (e) {
            await addGPXLayer(e.target?.result as string);
            const closestCity = (
                await ms.index("cities500").search("", {
                    filter: [`_geoRadius(${$form.lat}, ${$form.lon}, 10000)`],
                    sort: [`_geoPoint(${$form.lat}, ${$form.lon}):asc`],
                    limit: 1,
                })
            ).hits[0];

            $form.location = closestCity.name;
        };
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
            $form.expand.waypoints.splice(index, 1);
            $form.expand.waypoints = $form.expand.waypoints;
        }
    }

    function beforeWaypointModalOpen() {
        const mapCenter = map.getCenter();
        waypoint.set(new Waypoint(mapCenter.lat, mapCenter.lng));
        openWaypointModal();
    }

    function saveWaypoint(e: CustomEvent<Waypoint>) {
        const savedWaypoint = e.detail;
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
        const marker = createMarkerFromWaypoint(L, savedWaypoint);
        marker.addTo(map);
        savedWaypoint.marker = marker;
    }

    function openPhotoBrowser() {
        document.getElementById("photoInput")!.click();
    }

    function handlePhotoSelection() {
        const files = (
            document.getElementById("photoInput") as HTMLInputElement
        ).files;

        if (!files) {
            return;
        }

        for (const file of files) {
            $form._photoFiles.push(file);

            (function (file) {
                var reader = new FileReader();
                reader.onload = function (e) {
                    if (e.target?.result) {
                        if ($form.photos.length == 0) {
                            $form.thumbnail = e.target.result as string;
                        }
                        $form.photos = [
                            ...$form.photos,
                            e.target.result as string,
                        ];
                    }
                };
                reader.readAsDataURL(file);
            })(file);
        }
    }

    function makePhotoThumbnail(index: number) {
        $form.thumbnail = $form.photos[index];
    }

    function handlePhotoDelete(index: number) {
        const photoToDelete = $form.photos[index];
        let reassignThumbnail: boolean = false;
        if ($form.thumbnail == photoToDelete) {
            reassignThumbnail = true;
        }
        $form.photos.splice(index, 1);
        $form._photoFiles.splice(index, 1);
        $form.photos = $form.photos;
        if (reassignThumbnail) {
            $form.thumbnail = $form.photos[0];
        }
    }

    function beforeSummitLogModalOpen() {
        summitLog.set(new SummitLog(format(new Date(), "yyyy-MM-dd")));
        openSummitLogModal();
    }

    function saveSummitLog(e: CustomEvent<SummitLog>) {
        const savedSummitLog = e.detail;
        savedSummitLog.date = format(savedSummitLog.date, "yyyy-MM-dd");
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
            currentSummitLog.date = format(currentSummitLog.date, "yyyy-MM-dd");
            summitLog.set(currentSummitLog);
            openSummitLogModal();
        } else if (e.detail.value === "delete") {
            $form.expand.summit_logs.splice(index, 1);
            $form.expand.summit_logs = $form.expand.summit_logs;
        }
    }
</script>

<main class="grid grid-cols-1 md:grid-cols-[400px_1fr]">
    <form
        id="trail-form"
        class="overflow-y-auto overflow-x-hidden flex flex-col gap-4 px-8 order-1 md:order-none mt-8 md:mt-0"
        on:submit={handleSubmit}
    >
        <h3 class="text-xl font-semibold">Pick a trail</h3>
        <button class="btn-primary" type="button" on:click={openFileBrowser}
            >Upload GPX</button
        >
        <input
            type="file"
            name="gpx"
            id="fileInput"
            accept=".gpx"
            style="display: none;"
            on:change={handleFileSelection}
        />
        <hr class="border-separator" />
        <h3 class="text-xl font-semibold">Basic Info</h3>
        <div class="flex gap-4 justify-around">
            <div class="flex flex-col items-center">
                <span>Distance</span>
                <span class="font-medium">{formatMeters($form.distance)}</span>
                <input type="hidden" name="distance" value={$form.distance} />
            </div>
            <div class="flex flex-col items-center">
                <span>Elevation gain</span>
                <span class="font-medium"
                    >{formatMeters($form.elevation_gain)}</span
                >
                <input
                    type="hidden"
                    name="elevation_gain"
                    value={$form.elevation_gain}
                />
            </div>
            <div class="flex flex-col items-center">
                <span>Est. duration</span>
                <span class="font-medium">{formatTimeHHMM($form.duration)}</span
                >
                <input type="hidden" name="duration" value={$form.duration} />
            </div>
            <input type="hidden" name="lat" value={$form.lat} />
            <input type="hidden" name="lon" value={$form.lon} />
        </div>
        <TextField
            name="name"
            label="Name"
            on:change={handleChange}
            error={$errors.name}
            bind:value={$form.name}
        ></TextField>
        <TextField
            name="location"
            label="Location"
            error={$errors.location}
            bind:value={$form.location}
        ></TextField>
        <Textarea
            name="description"
            label="Describe your trail"
            bind:value={$form.description}
        ></Textarea>
        {#if $form.expand.category}
            <Select
                name="category"
                label="Category"
                bind:value={$form.expand.category.id}
                items={$categories.map((c) => ({ text: c.name, value: c.id }))}
            ></Select>
        {/if}
        <Toggle name="public" label="Public" bind:value={$form.public}></Toggle>
        <hr class="border-separator" />
        <h3 class="text-xl font-semibold">Waypoints</h3>
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
            ><i class="fa fa-plus mr-2"></i>Add Waypoint</button
        >
        <hr class="border-separator" />
        <h3 class="text-xl font-semibold">Photos</h3>
        <div class="flex gap-4 max-w-full overflow-x-auto shrink-0">
            <button
                class="btn-secondary h-32 w-32 m-2 shrink-0 grow-0 basis-auto"
                type="button"
                on:click={openPhotoBrowser}><i class="fa fa-plus"></i></button
            >
            <input
                type="file"
                id="photoInput"
                accept="image/*"
                multiple={true}
                style="display: none;"
                on:change={handlePhotoSelection}
            />
            {#each $form.photos ?? [] as photo, i}
                <div class="shrink-0 grow-0 basis-auto m-2">
                    <PhotoCard
                        src={photo}
                        on:delete={() => handlePhotoDelete(i)}
                        isThumbnail={$form.thumbnail === photo}
                        on:thumbnail={() => makePhotoThumbnail(i)}
                    ></PhotoCard>
                </div>
            {/each}
        </div>
        <hr class="border-separator" />
        <h3 class="text-xl font-semibold">Summit Book</h3>
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
            ><i class="fa fa-plus mr-2"></i>Add Entry</button
        >
        <hr class="border-separator" />
        <Button
            primary={true}
            large={true}
            type="submit"
            extraClasses="mb-2"
            {loading}>Save Trail</Button
        >
    </form>
    <div class="rounded-xl" id="map"></div>
</main>
<WaypointModal bind:openModal={openWaypointModal} on:save={saveWaypoint}
></WaypointModal>
<SummitLogModal bind:openModal={openSummitLogModal} on:save={saveSummitLog}
></SummitLogModal>

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
