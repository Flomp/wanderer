<script lang="ts">
    import Button from "$lib/components/base/button.svelte";
    import Select from "$lib/components/base/select.svelte";
    import TextField from "$lib/components/base/text_field.svelte";
    import Textarea from "$lib/components/base/textarea.svelte";
    import PhotoCard from "$lib/components/photo_card.svelte";
    import SummitLogCard from "$lib/components/summit_log/summit_log_card.svelte";
    import SummitLogModal from "$lib/components/summit_log/summit_log_modal.svelte";
    import WaypointCard from "$lib/components/waypoint/waypoint_card.svelte";
    import WaypointModal from "$lib/components/waypoint/waypoint_modal.svelte";
    import { pb } from "$lib/constants";
    import { SummitLog } from "$lib/models/summit_log";
    import { Trail, trailSchema } from "$lib/models/trail";
    import { Waypoint } from "$lib/models/waypoint";
    import { categories } from "$lib/stores/category_store";
    import { summitLog, summit_logs_create } from "$lib/stores/summit_log_store";
    import { trail, trails_create, trails_update } from "$lib/stores/trail_store";
    import { waypoint, waypoints_create } from "$lib/stores/waypoint_store";
    import { formatMeters, formatTimeHHMM } from "$lib/util/format_util";
    import { createMarkerFromWaypoint } from "$lib/util/leaflet_util";
    import { createForm } from "$lib/vendor/svelte-form-lib";
    import cryptoRandomString from "crypto-random-string";
    import { format } from "date-fns";
    import type { GPX, Icon, LeafletEvent, Map } from "leaflet";
    import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
    import "leaflet/dist/leaflet.css";
    import { onMount } from "svelte";

    let L: any;
    let map: Map;

    let gpxLayer: GPX;

    let openWaypointModal: () => void;
    let openSummitLogModal: () => void;

    let loading = false;
    const { form, errors, handleChange, handleSubmit } = createForm<Trail>({
        initialValues: new Trail("adsf", { category: $categories[0] }),
        validationSchema: trailSchema,
        onSubmit: async (trail) => {
            loading = true;
            try {
                const form = document.getElementById(
                    "trail-form",
                ) as HTMLFormElement;
                const formData = new FormData(form);
                formData.set("category", trail.expand.category!.id);

                for (const file of trail._photoFiles) {
                    formData.append("photos", file);
                }

                for (const waypoint of trail.expand.waypoints) {
                    const model = await waypoints_create({...waypoint, marker: undefined});
                    formData.append("waypoints", model.id!);
                }
                for (const summitLog of trail.expand.summit_logs) {
                    const model = await summit_logs_create(summitLog);
                    formData.append("summit_logs", model.id!);
                }
                const model = await trails_create(formData);

                const thumbnailIndex = trail.photos.findIndex(
                    (p) => p == trail.thumbnail,
                );
                let thumbnail: string | undefined = "/imgs/thumbnail.jpg";
                if (thumbnailIndex >= 0) {
                    thumbnail = model.photos.at(thumbnailIndex);
                }
                await trails_update(model.id!, { thumbnail: thumbnail });
            } catch (e) {
            } finally {
                loading = false;
            }
        },
    });

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

        $form.gpx = selectedFile?.name;
        $form.expand.waypoints = [];

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
                        $form.expand.waypoints.push(waypoint);
                    }
                })
                .on("loaded", function (e: LeafletEvent) {
                    map.fitBounds(e.target.getBounds());
                    $form.distance = e.target.get_distance();
                    $form.elevation_gain = e.target.get_elevation_gain();
                    $form.duration = e.target.get_total_time() / 1000 / 60;
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
            openWaypointModal();
        } else if (e.detail.value === "delete") {
            currentWaypoint.marker?.remove();
            $form.expand.waypoints = $form.expand.waypoints.filter(
                (w) =>
                    w.lat !== currentWaypoint.lat &&
                    w.lon !== currentWaypoint.lon,
            );
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
        summitLog.set(new SummitLog(new Date().toISOString()));
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
            summitLog.set(currentSummitLog);
            openSummitLogModal();
        } else if (e.detail.value === "delete") {
            $form.expand.summit_logs.splice(index, 1);
            $form.expand.summit_logs = $form.expand.summit_logs;
        }
    }
</script>

<main class="grid grid-cols-[400px_1fr]">
    <form
        id="trail-form"
        class="overflow-y-scroll overflow-x-hidden flex flex-col gap-4 px-8"
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
        <hr />
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
        </div>
        <TextField
            name="name"
            label="Name"
            error={$errors.name}
            bind:value={$form.name}
        ></TextField>
        <Textarea
            name="description"
            label="Describe your trail"
            bind:value={$form.description}
        ></Textarea>
        <Select
            name="category"
            label="Category"
            bind:value={$form.expand.category}
            items={$categories.map((c) => ({ text: c.name, value: c }))}
        ></Select>
        <hr />
        <h3 class="text-xl font-semibold">Waypoints</h3>
        <ul>
            {#each $form.expand.waypoints as waypoint, i}
                <li on:mouseenter={() => openMarkerPopup(waypoint)}>
                    <WaypointCard
                        {waypoint}
                        mode="edit"
                        on:change={(e) => handleWaypointMenuClick(waypoint, e)}
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
        <hr />
        <h3 class="text-xl font-semibold">Photos</h3>
        <div class="flex gap-4 max-w-full overflow-x-scroll shrink-0">
            <button
                class="btn-secondary h-32 w-32 shrink-0 grow-0 basis-auto"
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
            {#each $form.photos as photo, i}
                <div class="shrink-0 grow-0 basis-auto">
                    <PhotoCard
                        src={photo}
                        on:delete={() => handlePhotoDelete(i)}
                        isThumbnail={$form.thumbnail === photo}
                        on:thumbnail={() => makePhotoThumbnail(i)}
                    ></PhotoCard>
                </div>
            {/each}
        </div>
        <hr />
        <h3 class="text-xl font-semibold">Summit Book</h3>
        <ul>
            {#each $form.expand.summit_logs as log, i}
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
        <hr />
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
    #map,
    form {
        height: calc(100vh - 124px);
    }
</style>
