<script lang="ts">
    import { goto } from "$app/navigation";
    import type { List } from "$lib/models/list";
    import type { Trail } from "$lib/models/trail";
    import {
        lists_add_trail,
        lists_index,
        lists_remove_trail,
    } from "$lib/stores/list_store";
    import { show_toast } from "$lib/stores/toast_store";
    import { trails_delete } from "$lib/stores/trail_store";
    import { trail2gpx } from "$lib/util/gpx_util";
    import { gpx } from "$lib/vendor/toGeoJSON/toGeoJSON";
    import { _ } from "svelte-i18n";
    import Dropdown, { type DropdownItem } from "../base/dropdown.svelte";
    import ConfirmModal from "../confirm_modal.svelte";
    import ListSelectModal from "../list/list_select_modal.svelte";
    import TrailExportModal from "./trail_export_modal.svelte";
    import JSZip from "jszip";
    import { getFileURL, saveAs } from "$lib/util/file_util";
    import TrailShareModal from "./trail_share_modal.svelte";
    import { currentUser } from "$lib/stores/user_store";
    import { pb } from "$lib/pocketbase";

    export let trail: Trail;
    export let mode: "overview" | "map";

    let openConfirmModal: () => void;
    let openListSelectModal: () => void;
    let openExportModal: () => void;
    let openShareModal: () => void;

    const allowEdit =
        trail.author == $currentUser?.id ||
        trail.expand.trail_share_via_trail?.some((s) => s.permission == "edit");

    const dropdownItems: DropdownItem[] = [
        mode == "overview"
            ? { text: $_("show-on-map"), value: "show", icon: "map" }
            : {
                  text: $_("show-in-overview"),
                  value: "show",
                  icon: "table-columns",
              },

        { text: $_("directions"), value: "direction", icon: "car" },
        ...(trail.gpx
            ? [
                  {
                      text: $_("export"),
                      value: "download",
                      icon: "download",
                  },
              ]
            : []),
        { text: $_("print"), value: "print", icon: "print" },
        { text: $_("add-to-list"), value: "list", icon: "bookmark" },
        ...(trail.author != pb.authStore.model?.id
            ? []
            : [{ text: $_("share"), value: "share", icon: "share" }]),
        ...(allowEdit
            ? [{ text: $_("edit"), value: "edit", icon: "pen" }]
            : []),
        ...(trail.author == $currentUser?.id
            ? [{ text: $_("delete"), value: "delete", icon: "trash" }]
            : []),
    ];

    async function handleDropdownClick(item: { text: string; value: any }) {
        if (item.value == "show") {
            goto(
                mode == "overview"
                    ? `/map/trail/${trail.id!}`
                    : `/trail/view/${trail.id!}`,
            );
        } else if (item.value == "list") {
            openListSelectModal();
        } else if (item.value == "direction") {
            window
                .open(
                    `https://www.google.com/maps/dir/Current+Location/${trail.lat},${trail.lon}`,
                    "_blank",
                )
                ?.focus();
        } else if (item.value == "print") {
            goto(`/map/trail/${trail.id}/print`);
        } else if (item.value == "share") {
            openShareModal();
        } else if (item.value == "download") {
            openExportModal();
        } else if (item.value == "edit") {
            goto(`/trail/edit/${trail.id}`);
        } else if (item.value == "delete") {
            openConfirmModal();
        }
    }

    async function exportTrail(exportSettings: {
        fileFormat: "gpx" | "json";
        photos: boolean;
        summitLog: boolean;
    }) {
        try {
            let fileData: string = await trail2gpx(trail);
            if (exportSettings.fileFormat == "json") {
                fileData = JSON.stringify(
                    gpx(new DOMParser().parseFromString(fileData, "text/xml")),
                );
            }
            const zip = new JSZip();
            zip.file(`${trail.name}.${exportSettings.fileFormat}`, fileData);
            if (exportSettings.photos) {
                const photoFolder = zip.folder($_("photos"));
                for (const photo of trail.photos) {
                    const photoURL = getFileURL(trail, photo);
                    const photoBlob = await fetch(photoURL).then((response) =>
                        response.blob(),
                    );
                    const photoData = new File([photoBlob], photo);
                    photoFolder?.file(photo, photoData, { base64: true });
                }
            }
            if (exportSettings.summitLog) {
                let summitLogString = "";
                for (const summitLog of trail.expand.summit_logs) {
                    summitLogString += `${summitLog.date},${summitLog.text}\n`;
                }
                zip.file(
                    `${trail.name} - ${$_("summit-book")}.csv`,
                    summitLogString,
                );
            }
            const blob = await zip.generateAsync({ type: "blob" });
            saveAs(blob, `${trail.name}.zip`);
        } catch (e) {
            console.error(e);
            show_toast({
                type: "error",
                icon: "close",
                text: $_("error-exporting-trail"),
            });
        }
    }

    async function deleteTrail() {
        trails_delete(trail).then(() => history.back());
    }

    async function handleListSelection(list: List) {
        try {
            if (list.trails?.includes(trail.id!)) {
                await lists_remove_trail(list, trail);
                show_toast({
                    type: "success",
                    icon: "check",
                    text: `${$_("removed-trail-from")} "${list.name}"`,
                });
            } else {
                await lists_add_trail(list, trail);
                show_toast({
                    type: "success",
                    icon: "check",
                    text: `${$_("added-trail-to")} "${list.name}"`,
                });
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
</script>

<Dropdown
    items={dropdownItems}
    on:change={(e) => handleDropdownClick(e.detail)}
    let:toggleMenu={openDropdown}
    ><button
        class="rounded-full bg-white text-black hover:bg-gray-200 focus:ring-4 ring-gray-100/50 transition-colors h-12 w-12"
        on:click={openDropdown}
    >
        <i class="fa fa-ellipsis-vertical"></i>
    </button></Dropdown
>

<ConfirmModal
    text={$_("delete-trail-confirm")}
    bind:openModal={openConfirmModal}
    on:confirm={deleteTrail}
></ConfirmModal>
<ListSelectModal
    bind:openModal={openListSelectModal}
    on:change={(e) => handleListSelection(e.detail)}
></ListSelectModal>
<TrailExportModal
    bind:openModal={openExportModal}
    on:export={(e) => exportTrail(e.detail)}
></TrailExportModal>
<TrailShareModal {trail} bind:openShareModal></TrailShareModal>
