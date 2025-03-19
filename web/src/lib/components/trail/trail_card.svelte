<script lang="ts">
    import { createBubbler } from "svelte/legacy";

    const bubble = createBubbler();
    import emptyStateTrailDark from "$lib/assets/svgs/empty_states/empty_state_trail_dark.svg";
    import emptyStateTrailLight from "$lib/assets/svgs/empty_states/empty_state_trail_light.svg";
    import type { Trail } from "$lib/models/trail";
    import { pb } from "$lib/pocketbase";
    import { theme } from "$lib/stores/theme_store";
    import { getFileURL, isVideoURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import ShareInfo from "../share_info.svelte";
    import type { MouseEventHandler } from "svelte/elements";
    import Chip from "../base/chip.svelte";

    interface Props {
        trail: Trail;
        fullWidth?: boolean;
        onmouseenter?: MouseEventHandler<HTMLDivElement>;
        onmouseleave?: MouseEventHandler<HTMLDivElement>;
    }

    let {
        trail,
        fullWidth = false,
        onmouseenter,
        onmouseleave,
    }: Props = $props();

    let thumbnail = $derived(
        trail.photos.length
            ? getFileURL(trail, trail.photos[trail.thumbnail ?? 0])
            : $theme === "light"
              ? emptyStateTrailLight
              : emptyStateTrailDark,
    );

    let trailIsShared = $derived(
        (trail.expand?.trail_share_via_trail?.length ?? 0) > 0,
    );
</script>

<div
    class="trail-card relative rounded-2xl border border-input-border min-w-72 h-[386px] {fullWidth
        ? ''
        : 'lg:w-72'} cursor-pointer flex flex-col"
    {onmouseenter}
    {onmouseleave}
    role="listitem"
>
    <div
        class="relative w-full basis-full max-h-48 overflow-hidden rounded-t-2xl"
    >
        {#if isVideoURL(thumbnail)}
            <!-- svelte-ignore a11y_media_has_caption -->
            <video
                id="header-img"
                class="w-full h-full object-cover"
                autoplay
                loop
                src={thumbnail}
            ></video>
        {:else}
            <img
                loading="lazy"
                class="w-full h-full"
                id="header-img"
                src={thumbnail}
                alt=""
            />
        {/if}
    </div>
    {#if (trail.public || trailIsShared) && pb.authStore.record}
        <div
            class="flex absolute top-4 right-4 {trail.public && trailIsShared
                ? 'w-14'
                : 'w-8'} h-8 rounded-full items-center justify-center bg-background text-content"
        >
            {#if trail.public && pb.authStore.record}
                <span
                    class="tooltip"
                    class:mr-2={trail.public && trailIsShared}
                    data-title={$_("public")}
                >
                    <i class="fa fa-globe"></i>
                </span>
            {/if}
            {#if trail.expand?.trail_share_via_trail?.length}
                <span class="tooltip" data-title={$_("shared")}>
                    <i class="fa fa-share-nodes"></i>
                </span>
            {/if}
        </div>
    {/if}
    <div class="p-4">
        <div>
            <h4 class="font-semibold text-lg line-clamp-2">{trail.name}</h4>
            {#if trail.date}
                <p class="text-xs text-gray-500 mb-3">
                    {new Date(trail.date).toLocaleDateString(undefined, {
                        month: "long",
                        day: "2-digit",
                        year: "numeric",
                        timeZone: "UTC",
                    })}
                </p>
            {/if}
            {#if trail.expand?.author}
                <p class="text-xs text-gray-500 mb-3">
                    {$_("by")}
                    <img
                        class="rounded-full w-5 aspect-square mx-1 inline"
                        src={getFileURL(
                            trail.expand.author,
                            trail.expand.author.avatar,
                        ) ||
                            `https://api.dicebear.com/7.x/initials/svg?seed=${trail.expand.author.username}&backgroundType=gradientLinear`}
                        alt="avatar"
                    />
                    {trail.expand.author.username}
                </p>
            {/if}
            {#if trail.tags?.length}
                <div class="flex flex-wrap gap-1 mb-3">
                    {#each trail.tags ?? [] as t}
                        <Chip text={t} closable={false} primary={false}></Chip>
                    {/each}
                </div>
            {/if}
            <div class="flex gap-x-4">
                {#if trail.location}
                    <h5>
                        <i class="fa fa-location-dot mr-3"></i>{trail.location}
                    </h5>
                {/if}
                <h5>
                    <i class="fa fa-gauge mr-3"></i>{$_(
                        trail.difficulty ?? "?",
                    )}
                </h5>
            </div>
        </div>
        <div
            class="grid grid-cols-2 mt-2 gap-1 text-sm text-gray-500 whitespace-nowrap"
        >
            <span
                ><i class="fa fa-left-right mr-2"></i>{formatDistance(
                    trail.distance,
                )}</span
            >
            <span
                ><i class="fa fa-clock mr-2"></i>{formatTimeHHMM(
                    trail.duration,
                )}</span
            >
            <span
                ><i class="fa fa-arrow-trend-up mr-2"></i>{formatElevation(
                    trail.elevation_gain,
                )}</span
            >
            <span
                ><i class="fa fa-arrow-trend-down mr-2"></i>{formatElevation(
                    trail.elevation_loss,
                )}</span
            >
        </div>
    </div>
</div>

<style>
    .trail-card #header-img {
        object-fit: cover;
        transition: 0.25s ease;
    }

    .trail-card:hover #header-img {
        scale: 1.075;
    }
</style>
