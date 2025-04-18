<script lang="ts">
    import type { SummitLog } from "$lib/models/summit_log";

    import { getFileURL, isVideoURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import PhotoGallery from "../photo_gallery.svelte";

    interface Props {
        log: SummitLog;
        showCategory?: boolean;
        showTrail?: boolean;
        showRoute?: boolean;
        showAuthor?: boolean;
        showDescription?: boolean;
        showPhotos?: boolean;
        ontext?: (summitLog: SummitLog) => void;
        onopen?: (summitLog: SummitLog) => void;
    }

    let {
        log,
        showCategory = false,
        showTrail = false,
        showRoute = false,
        showAuthor = false,
        showDescription = false,
        showPhotos = false,
        onopen,
        ontext,
    }: Props = $props();

    let gallery: PhotoGallery;

    let imgSrc: string[] = $state([]);
    $effect(() => {
        if (log.photos?.length) {
            imgSrc = log.photos
                .filter((_, i) => i < 3)
                .reverse()
                .map((p) => getFileURL(log, p));
        } else {
            imgSrc = [];
        }
    });

    function openText() {
        ontext?.(log);
    }

    function openRoute() {
        onopen?.(log);
    }

    function colCount() {
        return [
            showCategory,
            showTrail,
            showAuthor,
            showRoute,
            showDescription,
        ].reduce((b, v) => (v ? b + 1 : b), 7);
    }
</script>

<tr class="text-sm">
    {#if showPhotos}
        <td>
            {#if imgSrc.length}
                <PhotoGallery
                    photos={log.photos.map((p) => getFileURL(log, p))}
                    bind:this={gallery}
                ></PhotoGallery>
                <button
                    class="relative w-16 aspect-square ml-2 mb-3 shrink-0"
                    type="button"
                    onclick={() => gallery.openGallery()}
                >
                    {#each imgSrc as img, i}
                        {#if isVideoURL(img)}
                            <!-- svelte-ignore a11y_media_has_caption -->
                            <video
                                controls={false}
                                loop
                                class="absolute h-full w-full rounded-xl object-cover"
                                style="top: {4 * i}px; right: {4 *
                                    i}px; transform: rotate(-{i * 5}deg)"
                                onmouseenter={(e) => (e.target as any).play()}
                                onmouseleave={(e) => (e.target as any).pause()}
                                src={img}
                            ></video>
                        {:else}
                            <img
                                class="absolute h-full rounded-xl object-cover"
                                style="top: {4 * i}px; right: {4 *
                                    i}px; transform: rotate(-{i * 5}deg)"
                                src={img}
                                alt="waypoint"
                            />
                        {/if}
                    {/each}
                </button>
            {/if}
        </td>
    {/if}
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
                aria-label="Go to trail"
                class="btn-icon aspect-square"
                href="/trail/view/{log.expand?.trails_via_summit_logs?.at(0)
                    ?.id ?? ''}"
                ><i class="fa fa-arrow-up-right-from-square px-[3px]"></i></a
            >
        </td>
    {/if}
    {#if showDescription}
        <td>
            {#if log.text}
                <button onclick={openText}
                    ><p
                        class="rounded-full bg-menu-item-background-hover hover:bg-menu-item-background-focus text-ellipsis max-w-28 whitespace-nowrap overflow-hidden px-3 py-1"
                    >
                        {log.text}
                    </p></button
                >
            {/if}
        </td>
    {/if}
    {#if showAuthor && log.expand?.author}
        <td>
            <p
                class="tooltip flex justify-center"
                data-title={log.expand.author.username}
            >
                {#if !log.expand.author.private}
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
                {:else}
                    <img
                        class="rounded-full w-7 aspect-square"
                        src={getFileURL(
                            log.expand?.author,
                            log.expand?.author.avatar,
                        ) ||
                            `https://api.dicebear.com/7.x/initials/svg?seed=${log.expand?.author.username}&backgroundType=gradientLinear`}
                        alt="avatar"
                    />
                {/if}
            </p>
        </td>
    {/if}
    {#if showRoute && log.gpx}
        <td>
            <button
                aria-label="Open route"
                onclick={openRoute}
                class="btn-icon"
            >
                <i class="fa fa-map-location-dot px-[3px] text-xl"></i></button
            >
        </td>
    {/if}
</tr>
<tr>
    <td colspan={colCount()}> <hr class="border-input-border" /> </td>
</tr>

<style>
    td {
        padding: 0.3rem 0.5rem;
    }
</style>
