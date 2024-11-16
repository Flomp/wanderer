<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import "leaflet/dist/leaflet.css";
    import { onMount, tick } from "svelte";
    import Modal from "../base/modal.svelte";
    import SummitLogTableRow from "./summit_log_table_row.svelte";

    import type { Trail } from "$lib/models/trail";
    import { gpx2trail } from "$lib/util/gpx_util";
    import type { Map } from "leaflet";
    import { _ } from "svelte-i18n";
    import MapWithElevation from "../trail/map_with_elevation.svelte";

    export let summitLogs: SummitLog[] = [];
    export let showCategory: boolean = false;
    export let showTrail: boolean = false;
    export let showAuthor: boolean = false;

    let openMapModal: () => void;
    let closeMapModal: () => void;

    let openTextModal: () => void;
    let closeTextModal: () => void;

    let map: Map;

    let trail: Trail | null = null;

    let currentText: string = "";

    onMount(async () => {});

    async function openMap(log: SummitLog) {
        if (!log.expand.gpx_data) {
            return;
        }

        trail = (await gpx2trail(log.expand.gpx_data)).trail;
        trail.expand.gpx_data = log.expand.gpx_data;

        openMapModal();
        await tick();

        map.invalidateSize();
        return;
    }

    async function openText(log: SummitLog) {
        currentText = log.text ?? "";
        openTextModal();
    }
</script>

<table class="w-full">
    <thead>
        <tr class="text-sm">
            <th class="w-24"></th>
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
            <th>{$_("description")}</th>
            {#if showAuthor}
                <th>
                    {$_("author", { values: { n: 1 } })}
                </th>
            {/if}
        </tr>
    </thead>
    <tbody>
        {#each summitLogs as log, i}
            <SummitLogTableRow
                index={i}
                {log}
                on:open={(e) => openMap(e.detail)}
                on:text={(e) => openText(e.detail)}
                {showCategory}
                {showTrail}
                {showAuthor}
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
        <MapWithElevation {trail} bind:map></MapWithElevation>
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
        padding: 0rem 0.75rem;
    }
</style>
