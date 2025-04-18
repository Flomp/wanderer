<script lang="ts">
    import type { Trail, TrailFilter } from "$lib/models/trail";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import type { SelectItem } from "../base/select.svelte";
    import { goto } from "$app/navigation";
    import { getFileURL } from "$lib/util/file_util";
    import ShareInfo from "../share_info.svelte";

    interface Props {
        tableHeader: SelectItem[];
        trails?: Trail[] | null;
        filter?: TrailFilter | null;
        onsort?: (value: any) => void
    }

    let { tableHeader, trails = null, filter = null, onsort }: Props = $props();

    function getColumnWidth(columnValue: string): string {
        switch (columnValue) {
            case "name":
                return "w-[25%]";
            case "distance":
                return "w-[15%]";
            case "duration":
            case "difficulty":
                return "w-[11%]";
            case "elevation_gain":
            case "created":
            case "date":
                return "w-[5%]";
            default:
                return "";
        }
    }
</script>

<div
    class="table-container w-full border border-input-border rounded-xl overflow-x-scroll overflow-y-clip"
>
    <table class="w-full">
        <thead>
            <tr class="bg-secondary-hover">
                {#each tableHeader as column}
                    <th
                        class="p-4 text-left text-sm font-medium {getColumnWidth(
                            column.value,
                        )}"
                        onclick={() => onsort?.(column.value)}
                    >
                        <div class="cursor-pointer">
                            {column.text}
                            {#if filter && filter.sort === column.value}
                                <i
                                    id="sort-order-btn"
                                    class="fa fa-arrow-up"
                                    class:rotated={filter.sortOrder == "-"}
                                ></i>
                            {/if}
                        </div>
                    </th>
                {/each}
            </tr>
        </thead>
        <tbody>
            {#if trails}
                {#each trails as trail}
                    <tr
                        class="border-t border-input-border cursor-pointer hover:bg-secondary-hover transition-colors"
                        onclick={() => goto(`/trail/view/${trail.id}`)}
                    >
                        <td
                            class="flex justify-between items-center text-sm relative"
                        >
                            <div class="p-4 w-[75%]" title={trail.name}>
                                {trail.name}
                            </div>
                            <div class="flex flex-col items-center">
                                {#if trail.expand && trail.expand.author}
                                    <div class="author-icon">
                                        <img
                                            title={`${trail.public ? $_("public") + " " : ""}${$_("by")} ${trail.expand.author.username}`}
                                            class="rounded-full w-5 aspect-square mx-1 inline"
                                            src={getFileURL(
                                                trail.expand.author,
                                                trail.expand.author.avatar,
                                            ) ||
                                                `https://api.dicebear.com/7.x/initials/svg?seed=${trail.expand.author.username}&backgroundType=gradientLinear`}
                                            alt="avatar"
                                        />
                                    </div>
                                {/if}
                                <div class="flex gap-x-1">
                                    {#if trail.public}
                                        <div
                                            class="public-icon tooltip"
                                            data-title={$_("public")}
                                        >
                                            <i class="fa fa-globe"></i>
                                        </div>
                                    {/if}
                                    {#if trail.expand?.trail_share_via_trail?.length}
                                        <ShareInfo type="trail" subject={trail}
                                        ></ShareInfo>
                                    {/if}
                                </div>
                            </div>
                        </td>
                        <td class="p-4 text-sm">
                            {formatDistance(trail.distance)}
                        </td>
                        <td class="p-4 text-sm">
                            {formatTimeHHMM(trail.duration)}
                        </td>
                        <td class="p-4 text-sm">
                            {$_(trail.difficulty ?? "easy")}
                        </td>
                        <td class="p-4 text-sm">
                            {formatElevation(trail.elevation_gain)}
                        </td>
                        <td class="p-4 text-sm">
                            {#if trail.created}
                                {new Date(trail.created).toLocaleDateString(
                                    undefined,
                                    {
                                        month: "2-digit",
                                        day: "2-digit",
                                        year: "numeric",
                                        timeZone: "UTC",
                                    },
                                )}
                            {/if}
                        </td>
                        <td class="p-4 text-sm">
                            {#if trail.date}
                                {new Date(trail.date).toLocaleDateString(
                                    undefined,
                                    {
                                        month: "2-digit",
                                        day: "2-digit",
                                        year: "numeric",
                                        timeZone: "UTC",
                                    },
                                )}
                            {/if}
                        </td>
                    </tr>
                {/each}
            {:else}
                {#each { length: 12 } as _, index}
                    <tr
                        class="border-t border-input-border cursor-pointer bg-secondary-hover transition-colors animate-pulse"
                    >
                        {#each { length: tableHeader.length } as _, index}
                            <td class="p-4 text-sm"
                                ><div
                                    class="h-4 bg-menu-item-background-focus rounded w-3/4"
                                ></div></td
                            >
                        {/each}
                    </tr>
                {/each}
            {/if}
        </tbody>
    </table>
</div>

<style>
    .table-container {
        container-type: inline-size;
    }

    @container (max-width: 660px) {
        th:nth-last-child(-n + 2),
        td:nth-last-child(-n + 2) {
            display: none;
        }
    }
</style>
