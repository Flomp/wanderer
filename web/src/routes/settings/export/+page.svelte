<script lang="ts">
    import Button from "$lib/components/base/button.svelte";
    import Settings from "$lib/components/profile/settings.svelte";
    import TrailExportModal from "$lib/components/trail/trail_export_modal.svelte";
    import { show_toast } from "$lib/stores/toast_store";
    import { fetchGPX, trails_index, trails_upload } from "$lib/stores/trail_store";
    import { getFileURL, saveAs } from "$lib/util/file_util";
    import { trail2gpx } from "$lib/util/gpx_util";
    import { gpx } from "$lib/vendor/toGeoJSON/toGeoJSON";
    import JSZip from "jszip";
    import { _ } from "svelte-i18n";
    import { linear } from "svelte/easing";
    import { tweened } from "svelte/motion";

    const uploadProgress = tweened(0, {
        duration: 300,
        easing: linear,
    });

    let openExportModal: () => void;

    function openFileBrowser() {
        document.getElementById("file-input")!.click();
    }

    function handleDragOver(e: DragEvent) {
        e.preventDefault();
    }

    function handleDrop(e: DragEvent) {
        e.preventDefault();
        handleFileSelection(e.dataTransfer?.files);
    }

    async function handleFileSelection(files?: FileList | null) {
        if (!files) {
            files = (document.getElementById("file-input") as HTMLInputElement)
                .files;
        }

        if (!files) {
            return;
        }

        let errorsThrown = 0;
        for (let i = 0; i < files.length; i++) {
            const file = files[i];
            try {
                await trails_upload(file);
            } catch (e) {
                errorsThrown += 1;
                show_toast({
                    type: "error",
                    icon: "close",
                    text: `Error uploading file: ${file.name}`,
                });
            }
            $uploadProgress += (files.length / i) * 100;
            setTimeout(() => {}, 3000);
        }
        if (errorsThrown == 0) {
            show_toast({
                type: "success",
                icon: "check",
                text: `${files.length} ${$_("trail", { values: { n: files.length } })} ${$_("uploaded")}`,
            });
        }
    }

    async function exportTrails(exportSettings: {
        fileFormat: "gpx" | "json";
        photos: boolean;
        summitLog: boolean;
    }) {
        try {
            const trails = await trails_index(-1);

            const zip = new JSZip();

            for (const trail of trails) {
                const gpxData: string = await fetchGPX(trail);
                if (!trail.expand) {
                    trail.expand = {};
                }
                trail.expand.gpx_data = gpxData;
                const trailFolder = zip.folder(`${trail.name}`);
                let fileData: string = await trail2gpx(trail);
                if (exportSettings.fileFormat == "json") {
                    fileData = JSON.stringify(
                        gpx(
                            new DOMParser().parseFromString(
                                fileData,
                                "text/xml",
                            ),
                        ),
                    );
                }
                trailFolder?.file(
                    `${trail.name}.${exportSettings.fileFormat}`,
                    fileData,
                );
                if (exportSettings.photos) {
                    const photoFolder = trailFolder?.folder($_("photos"));
                    for (const photo of trail.photos ?? []) {
                        const photoURL = getFileURL(trail, photo);
                        const photoBlob = await fetch(photoURL).then(
                            (response) => response.blob(),
                        );
                        const photoData = new File([photoBlob], photo);
                        photoFolder?.file(photo, photoData, { base64: true });
                    }
                }
                if (exportSettings.summitLog) {
                    let summitLogString = "";
                    for (const summitLog of trail.expand?.summit_logs ?? []) {
                        summitLogString += `${summitLog.date},${summitLog.text}\n`;
                    }
                    trailFolder?.file(
                        `${trail.name} - ${$_("summit-book")}.csv`,
                        summitLogString,
                    );
                }
            }

            const blob = await zip.generateAsync({ type: "blob" });
            saveAs(blob, `wanderer-export-${Date.now()}.zip`);
        } catch (e) {
            console.error(e);
            show_toast({
                type: "error",
                icon: "close",
                text: $_("error-exporting-trail"),
            });
        }
    }
</script>

<svelte:head>
    <title>{$_("settings")} | wanderer</title>
</svelte:head>
<main>
    <Settings selected={2}>
        <div class="space-y-6">
            <h3 class="text-2xl font-semibold">{$_("import")}</h3>
            <button
                class="drop-area relative h-64 w-full p-4 border border-content border-dashed rounded-xl flex items-center justify-center text-gray-500 bg-background cursor-pointer hover:bg-menu-item-background-hover focus:bg-menu-item-background-focus transition-colors"
                style="--progress: {$uploadProgress}%"
                on:click={openFileBrowser}
                on:dragover={handleDragOver}
                on:drop={handleDrop}
            >
                {$_('import-hint')}
            </button>

            <input
                type="file"
                id="file-input"
                accept=".gpx,.GPX,.tcx,.TCX,.kml,.KML,.fit,.FIT"
                multiple={true}
                style="display: none;"
                on:change={() => handleFileSelection()}
            />
            <h3 class="text-2xl font-semibold">{$_("export")}</h3>
            <Button secondary={true} on:click={() => openExportModal()}
                >{$_("export-all-trails")}</Button
            >
        </div>
    </Settings>
</main>

<TrailExportModal
    bind:openModal={openExportModal}
    on:export={(e) => exportTrails(e.detail)}
></TrailExportModal>

<style>
    .drop-area::after {
        content: "";
        display: block;
        position: absolute;
        left: -4px;
        top: -4px;
        right: -4px;
        bottom: -4px;
        background-color: transparent;
        z-index: -100;
        border-radius: 0.75rem;
        background-image: conic-gradient(
            rgba(var(--content)),
            rgba(var(--content)) var(--progress),
            transparent var(--progress)
        );
    }
</style>
