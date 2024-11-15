<script lang="ts">
    import type { Trail } from "$lib/models/trail";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import ShareInfo from "../share_info.svelte";
    import { pb } from "$lib/pocketbase";

    export let trail: Trail;

    export let showDescription: boolean = true;

    $: thumbnail = trail.photos.length
        ? getFileURL(trail, trail.photos[trail.thumbnail])
        : "/imgs/default_thumbnail.webp";
</script>

<li
    class="flex gap-8 p-4 rounded-xl border border-input-border cursor-pointer hover:bg-secondary-hover transition-colors items-center"
>
    <div class="shrink-0">
        <img class="h-28 w-28 object-cover rounded-xl" src={thumbnail} alt="" />
    </div>
    <div class="min-w-0 basis-full">
        <div class="flex items-center gap-4">
            <h4 class="font-semibold text-lg">
                {trail.name}
            </h4>
            {#if trail.public && pb.authStore.model}
                <span class="tooltip ml-3" data-title={$_("public")}>
                    <i class="fa fa-globe"></i>
                </span>
            {/if}
            {#if trail.expand?.trail_share_via_trail?.length}
                <ShareInfo type="trail" subject={trail}></ShareInfo>
            {/if}
        </div>
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
        {#if trail.expand.author}
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
        <div class="flex flex-wrap gap-x-8">
            {#if trail.location}
                <h5><i class="fa fa-location-dot mr-3"></i>{trail.location}</h5>
            {/if}
            <h5>
                <i class="fa fa-gauge mr-3"></i>{$_(trail.difficulty ?? "?")}
            </h5>
        </div>

        <div class="flex flex-wrap mt-1 gap-x-4 gap-y-2 text-sm text-gray-500">
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
        {#if showDescription}
            <p
                class="mt-3 text-sm whitespace-nowrap min-w-0 max-w-full overflow-hidden text-ellipsis"
            >
                {trail.description}
            </p>
        {/if}
    </div>
</li>
