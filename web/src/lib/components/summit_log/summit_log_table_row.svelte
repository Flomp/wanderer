<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";
    import "leaflet/dist/leaflet.css";

    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { gpx } from "$lib/vendor/toGeoJSON/toGeoJSON";
    import type { Map } from "leaflet";
    import { createEventDispatcher, onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import { getFileURL } from "$lib/util/file_util";

    export let index: number;
    export let log: SummitLog;
    export let showCategory: boolean = false;
    export let showTrail: boolean = false;
    export let showAuthor: boolean = false;

    let map: Map;

    const dispatch = createEventDispatcher();

    onMount(async () => {
        if (log.expand.gpx_data) {
            await initMap();
        }
    });

    async function initMap() {
        const L = (await import("leaflet")).default;

        map = L.map("mini-map-" + index, {
            zoomControl: false,
            scrollWheelZoom: false,
            dragging: false,
        });
        map.attributionControl.setPrefix(false);

        if (!log.expand.gpx_data || !map) {
            return;
        }

        const geoJson = gpx(
            new DOMParser().parseFromString(log.expand.gpx_data, "text/xml"),
        );
        const layer = (L as any)
            .geoJson(geoJson, {
                filter: (feature: any, layer: any) => {
                    return feature.geometry.type !== "Point";
                },
            })
            .addTo(map);
        map.fitBounds(layer.getBounds());
        map.invalidateSize();
    }

    function openMap() {
        dispatch("open", log);
    }

    function openText() {
        dispatch("text", log);
    }

    function colCount() {
        return [showCategory, showTrail, showAuthor].reduce(
            (b, v) => (v ? b + 1 : b),
            7,
        );
    }
</script>

<tr class="text-center">
    <td>
        <button
            type="button"
            on:click={openMap}
            class="h-20 aspect-square shrink-0 rounded-xl !bg-background hover:!bg-secondary-hover transition-colors"
            class:hidden={!log.expand?.gpx_data}
            id="mini-map-{index}"
        ></button>
    </td>
    <td class:py-4={!log.expand?.gpx_data}
        >{new Date(log.date).toLocaleDateString(undefined, {
            month: "2-digit",
            day: "2-digit",
            year: "numeric",
            timeZone: "UTC",
        })}</td
    >
    <td>
        {formatDistance(log.distance)}
    </td>

    <td>
        {formatElevation(log.elevation_gain)}
    </td>
    <td>
        {formatElevation(log.elevation_loss)}
    </td>
    <td>
        {formatTimeHHMM(log.duration ? log.duration / 60 : undefined)}
    </td>
    {#if showCategory}
        <td>
            {$_(
                log.expand?.trails_via_summit_logs?.at(0)?.expand?.category
                    ?.name ?? "-",
            )}
        </td>
    {/if}
    {#if showTrail}
        <td>
            <a
                class="btn-icon aspect-square"
                href="/trail/view/{log.expand?.trails_via_summit_logs?.at(0)
                    ?.id ?? ''}"
                ><i class="fa fa-arrow-up-right-from-square px-[3px]"></i></a
            >
        </td>
    {/if}
    <td>
        {#if log.text}
            <button on:click={openText} class="btn-icon"
                ><i class="fa-regular fa-message"></i></button
            >
        {/if}
    </td>
    {#if showAuthor && log.expand.author}
        <td>
            <p
                class="tooltip flex justify-center"
                data-title={log.expand.author.username}
            >
                <a href="/profile/{log.expand.author.id}">
                    <img
                        class="rounded-full w-7 aspect-square"
                        src={getFileURL(
                            log.expand?.author,
                            log.expand?.author.avatar,
                        ) ||
                            `https://api.dicebear.com/7.x/initials/svg?seed=${log.expand?.author.username}&backgroundType=gradientLinear`}
                        alt="avatar"
                    />
                </a>
            </p>
        </td>
    {/if}
</tr>
<tr>
    <td colspan={colCount()}> <hr class="border-input-border" /> </td>
</tr>
