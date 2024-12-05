<script lang="ts">
    import type { Trail, TrailFilter } from "$lib/models/trail";
    import { formatDistance, formatElevation } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import type { SelectItem } from "../base/select.svelte";
    import { goto } from "$app/navigation";
    import { createEventDispatcher } from "svelte";

    export let tableHeader: SelectItem[];
    export let trails: Trail[] | null = null;
    export let filter: TrailFilter | null = null;

    let dispatch = createEventDispatcher();

    function getColumnWidth(columnValue: string): string {
        switch (columnValue) {
            case "name":
                return "w-[25%]";
            case "distance":
            case "difficulty":
                return "w-[10%] md-hidden";
            case "elevation_gain":
            case "elevation_loss":
                return "w-[10%]";
            case "created":
            case "date":
                return "w-[10%]";
            default:
                return "w-1/6";
        }
    }
</script>

<div
    class="table-container w-full border border-input-border rounded-xl overflow-hidden"
>
    <table class="w-full">
        <thead>
            <tr class="bg-secondary">
                {#each tableHeader as column}
                    <th
                        class="p-4 text-left text-sm font-medium {getColumnWidth(
                            column.value,
                        )}"
                        on:click={() => dispatch("sort", column.value)}
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
                        class="border-t border-input-border cursor-pointer hover:bg-secondary-hover transition-colors overflow-hidden"
                        on:click={() => goto(`/trail/view/${trail.id}`)}
                    >
                        <td class="p-4 text-sm line-clamp-2">
                            {trail.name}
                        </td>
                        <td class="p-4 text-sm">
                            {formatDistance(trail.distance)}
                        </td>
                        <td class="p-4 text-sm">
                            {trail.difficulty}
                        </td>
                        <td class="p-4 text-sm">
                            {formatElevation(trail.elevation_gain)}
                        </td>
                        <td class="p-4 text-sm">
                            {formatElevation(trail.elevation_loss)}
                        </td>
                        <td class="p-4 text-sm">
                            {#if trail.created}
                                {new Date(trail.created).toLocaleDateString(
                                    undefined,
                                    {
                                        month: "short",
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
                                        month: "short",
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

    @container (max-width: 550px) {
        th:nth-last-child(-n + 2),
        td:nth-last-child(-n + 2) {
            display: none;
        }
    }
</style>
