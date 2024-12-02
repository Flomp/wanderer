<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import { onMount, tick } from "svelte";
    import Modal from "../base/modal.svelte";
    import SummitLogTableRow from "./summit_log_table_row.svelte";

    import type { Trail } from "$lib/models/trail";
    import { gpx2trail } from "$lib/util/gpx_util";
    import * as M from "maplibre-gl";
    import { _ } from "svelte-i18n";
    import MapWithElevationMaplibre from "../trail/map_with_elevation_maplibre.svelte";

    export let summitLogs: SummitLog[] = [];
    export let showCategory: boolean = false;
    export let showTrail: boolean = false;
    export let showAuthor: boolean = false;
    export let showRoute: boolean = false;
    export let showPhotos: boolean = false;

    let openMapModal: () => void;
    let closeMapModal: () => void;

    let openTextModal: () => void;
    let closeTextModal: () => void;

    let map: M.Map;

    let trail: Trail | null = null;

    let currentText: string = "";

    onMount(async () => {});

    async function openMap(log: SummitLog) {
        if (!log.expand.gpx_data) {
            return;
        }

        trail = (await gpx2trail(log.expand.gpx_data)).trail;
        trail.id = log.id;
        trail.expand.gpx_data = log.expand.gpx_data;

        openMapModal();
        await tick();
        return;
    }

    async function openText(log: SummitLog) {
        currentText = log.text ?? "";
        openTextModal();
    }
</script>

<table class="w-full">
    <thead class="text-left text-gray-500">
        <tr class="text-sm">
            {#if showPhotos}
                <th class="w-24"></th>
            {/if}
            <th>{$_("date")}</th>
            <th>{$_("distance")}</th>
            <th>{$_("elevation-gain")}</th>
            <th>{$_("elevation-loss")}</th>
            <th>{$_("duration")}</th>
            {#if showCategory}
                <th>
                    {$_("category")}
                </th>
            {/if}
            {#if showTrail}
                <th>
                    {$_("trail", { values: { n: 1 } })}
                </th>
            {/if}
            {#if summitLogs.some((l) => l.text?.length)}
                <th>{$_("description")}</th>
            {/if}
            {#if showAuthor}
                <th>
                    {$_("author", { values: { n: 1 } })}
                </th>
            {/if}
            {#if showRoute && summitLogs.some((l) => l.gpx)}
                <th>
                    {$_("map")}
                </th>
            {/if}
        </tr>
    </thead>
    <tbody>
        {#each summitLogs as log, i}
            <SummitLogTableRow
                {log}
                on:open={(e) => openMap(e.detail)}
                on:text={(e) => openText(e.detail)}
                {showCategory}
                {showTrail}
                {showAuthor}
                {showPhotos}
                showDescription={summitLogs.some(l => l.text?.length)}
                showRoute={showRoute && summitLogs.some((l) => l.gpx)}
            ></SummitLogTableRow>
        {/each}
    </tbody>
</table>
{#if !summitLogs.length}
    <p class="text-center w-full my-8 text-gray-500 text-sm">{$_("no-data")}</p>
{/if}
<Modal
    id="summit-log-table-map-modal"
    size="max-w-4xl"
    title=""
    bind:openModal={openMapModal}
    bind:closeModal={closeMapModal}
>
    <div slot="content" id="summit-log-table-map" class="h-[32rem]">
        {#if trail}
            <MapWithElevationMaplibre trails={[trail]} bind:map showTerrain
            ></MapWithElevationMaplibre>
        {/if}
    </div>
</Modal>
<Modal
    id="summit-log-table-text-modal"
    size="max-w-xl"
    title={$_("description")}
    bind:openModal={openTextModal}
    bind:closeModal={closeTextModal}
>
    <p slot="content" class="whitespace-pre-wrap">{currentText}</p>
</Modal>

<style>
    th {
        padding: 0rem 0.5rem;
    }
</style>
