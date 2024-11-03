<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import "leaflet/dist/leaflet.css";
    import { onMount, tick } from "svelte";
    import Modal from "../base/modal.svelte";
    import SummitLogTableRow from "./summit_log_table_row.svelte";

    import type { Map } from "leaflet";
    import { gpx } from "$lib/vendor/toGeoJSON/toGeoJSON";
    import { endIcon, startIcon } from "$lib/util/leaflet_util";
    import { _ } from "svelte-i18n";
    import MapWithElevation from "../trail/map_with_elevation.svelte";
    import type { Trail } from "$lib/models/trail";
    import { gpx2trail } from "$lib/util/gpx_util";

    export let summitLogs: SummitLog[];
    export let showCategory: boolean = false;

    let openModal: () => void;
    let closeModal: () => void;

    let map: Map;

    let trail: Trail | null = null;

    onMount(async () => {
        // L = (await import("leaflet")).default;
        // map = L.map("summit-log-table-map");
        // map.attributionControl.setPrefix(false);
        // L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        //     attribution: "© OpenStreetMap contributors",
        // }).addTo(map);
        // layerGroup = L.layerGroup();
        // layerGroup.addTo(map);
    });

    async function openMap(log: SummitLog) {
        if (!log.expand.gpx_data) {
            return;
        }

        trail = (await gpx2trail(log.expand.gpx_data)).trail;
        trail.expand.gpx_data = log.expand.gpx_data;

        openModal();
        await tick();

        map.invalidateSize();
        return;
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
            <th></th>
        </tr>
    </thead>
    <tbody>
        {#each summitLogs as log, i}
            <SummitLogTableRow
                index={i}
                {log}
                on:open={() => openMap(log)}
                {showCategory}
            ></SummitLogTableRow>
        {/each}
    </tbody>
</table>

<Modal
    id="summit-log-table-modal"
    size="max-w-4xl"
    title=""
    bind:openModal
    bind:closeModal
>
    <div slot="content" id="summit-log-table-map" class="h-[32rem]">
        <MapWithElevation {trail} bind:map></MapWithElevation>
    </div>
</Modal>

<style>
    th {
        padding: 0rem 0.75rem;
    }
</style>
