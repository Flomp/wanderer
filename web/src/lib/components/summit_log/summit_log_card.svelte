<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import { _ } from "svelte-i18n";
    import Dropdown, { type DropdownItem } from "../base/dropdown.svelte";
    import emptyStateTrailDark from "$lib/assets/svgs/empty_states/empty_state_trail_dark.svg";
    import emptyStateTrailLight from "$lib/assets/svgs/empty_states/empty_state_trail_light.svg";

    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { getFileURL, readAsDataURLAsync } from "$lib/util/file_util";
    import { theme } from "$lib/stores/theme_store";

    let thumbnail: string =
        $theme === "light" ? emptyStateTrailLight : emptyStateTrailDark;

    $: Promise.all(
        (log._photos ?? []).map(async (f) => {
            return await readAsDataURLAsync(f);
        }),
    ).then((v) => {
        if (log.photos.length) {
            thumbnail = getFileURL(log, log.photos[0]);
        } else if (v.length) {
            thumbnail = v[0];
        } else {
            thumbnail =
                $theme === "light" ? emptyStateTrailLight : emptyStateTrailDark;
        }
    });

    export let log: SummitLog;
    export let mode: "show" | "edit" = "show";

    const dropdownItems: DropdownItem[] = [
        { text: $_("edit"), value: "edit" },
        { text: $_("delete"), value: "delete" },
    ];
</script>

<div class="p-4 my-2 border border-input-border rounded-xl">
    <div class="flex items-center gap-x-4">
        <div class="h-24 aspect-square shrink-0 rounded-xl overflow-hidden">
            <img
                id="header-img"
                class="object-cover h-full"
                src={thumbnail}
                alt=""
            />
        </div>
        <div class="basis-full">
            <div
                class="flex justify-between items-center"
                class:mb-2={log.text}
            >
                <h5 class="font-medium mr-2">
                    {new Date(log.date).toLocaleDateString(undefined, {
                        month: "2-digit",
                        day: "2-digit",
                        year: "numeric",
                        timeZone: "UTC",
                    })}
                </h5>

                {#if mode == "edit"}
                    <Dropdown items={dropdownItems} on:change></Dropdown>
                {/if}
            </div>
            {#if log.distance || log.elevation_gain || log.elevation_loss || log.duration}
                <div
                    class="flex mt-1 gap-x-4 text-sm text-gray-500 flex-wrap mb-2"
                >
                    <span
                        ><i class="fa fa-left-right mr-2"></i>{formatDistance(
                            log.distance,
                        )}</span
                    >
                    <span
                        ><i class="fa fa-clock mr-2"></i>{formatTimeHHMM(
                            log.duration ? log.duration / 60 : undefined,
                        )}</span
                    >
                    <span
                        ><i class="fa fa-arrow-trend-up mr-2"
                        ></i>{formatElevation(log.elevation_gain)}</span
                    >
                    <span
                        ><i class="fa fa-arrow-trend-down mr-2"
                        ></i>{formatElevation(log.elevation_loss)}</span
                    >
                </div>
            {/if}
            <span class="whitespace-pre-wrap">{log.text}</span>
        </div>
    </div>
</div>
