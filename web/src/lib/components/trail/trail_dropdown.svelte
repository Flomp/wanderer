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
    import { getFileURL } from "$lib/util/file_util";
    import Dropdown, { type DropdownItem } from "../base/dropdown.svelte";
    import ConfirmModal from "../confirm_modal.svelte";
    import ListSelectModal from "../list/list_select_modal.svelte";
    import { _ } from "svelte-i18n";

    export let trail: Trail;
    export let mode: "overview" | "map";

    let openConfirmModal: () => void;
    let openListSelectModal: () => void;

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
                      text: $_("download-gpx"),
                      value: "download",
                      icon: "download",
                  },
              ]
            : []),
        { text: $_("add-to-list"), value: "list", icon: "bookmark" },
        { text: $_("edit"), value: "edit", icon: "pen" },
        { text: $_("delete"), value: "delete", icon: "trash" },
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
        } else if (item.value == "download") {
            downloadURI(getFileURL(trail, trail.gpx), trail.gpx!);
        } else if (item.value == "edit") {
            goto(`/trail/edit/${trail.id}`);
        } else if (item.value == "delete") {
            openConfirmModal();
        }
    }

    function downloadURI(uri: string, name: string) {
        var link = document.createElement("a");
        link.setAttribute("download", name);
        link.href = uri;
        document.body.appendChild(link);
        link.click();
        link.remove();
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
