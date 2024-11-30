<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";

    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { createEventDispatcher } from "svelte";
    import { _ } from "svelte-i18n";
    import PhotoGallery from "../photo_gallery.svelte";

    export let log: SummitLog;
    export let showCategory: boolean = false;
    export let showTrail: boolean = false;
    export let showAuthor: boolean = false;

    let openGallery: (idx?: number) => void;

    let imgSrc: string[] = [];
    $: if (log.photos?.length) {
        imgSrc = log.photos
            .filter((_, i) => i < 3)
            .reverse()
            .map((p) => getFileURL(log, p));
    } else {
        imgSrc = [];
    }

    const dispatch = createEventDispatcher();

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
        {#if imgSrc.length}
            <PhotoGallery
                photos={log.photos.map((p) => getFileURL(log, p))}
                bind:open={openGallery}
            ></PhotoGallery>
            <button
                class="relative w-16 aspect-square ml-2 mb-3 shrink-0"
                type="button"
                on:click={() => openGallery()}
            >
                {#each imgSrc as img, i}
                    <img
                        class="absolute h-full rounded-xl object-cover"
                        style="top: {6 * i}px; right: {6 *
                            i}px; transform: rotate(-{i * 5}deg)"
                        src={img}
                        alt="waypoint"
                    />
                {/each}
            </button>
        {/if}
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
