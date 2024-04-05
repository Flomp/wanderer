<script lang="ts">
    import type { Trail } from "$lib/models/trail";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";

    export let trail: Trail;

    $: thumbnail = trail.photos.length
        ? getFileURL(trail, trail.photos[trail.thumbnail])
        : "/imgs/default_thumbnail.webp";
</script>

<div
    class="trail-card relative rounded-2xl border border-input-border sm:w-72 cursor-pointer"
    on:mouseenter
    on:mouseleave
    role="listitem"
>
    {#if trail.public}
        <div
            class=" top-4 right-4 z-10 tooltip"
            style="position: absolute!important"
            data-title={$_("public")}
        >
            <i class="fa fa-globe"></i>
        </div>
    {/if}
    <div
        class="relative w-full min-h-40 max-h-48 overflow-hidden rounded-t-2xl"
    >
        <img src={thumbnail} alt="" />
    </div>
    <div class="p-4">
        <div>
            <h4 class="font-semibold text-lg">{trail.name}</h4>
            {#if trail.date}
                <p class="text-xs text-gray-500 mb-3">
                    {new Date(trail.date).toLocaleDateString(undefined, {
                        month: "long",
                        day: "2-digit",
                        year: "numeric",
                        timeZone: 'UTC'
                    })}
                </p>
            {/if}
            <div class="flex gap-x-4">
                {#if trail.location}
                    <h5>
                        <i class="fa fa-location-dot mr-3"></i>{trail.location}
                    </h5>
                {/if}
                <h5>
                    <i class="fa fa-person-hiking mr-3"></i>{$_(
                        trail.difficulty ?? "?",
                    )}
                </h5>
            </div>
        </div>
        <div class="flex mt-1 gap-4 text-sm text-gray-500 whitespace-nowrap">
            <span
                ><i class="fa fa-left-right mr-2"></i>{formatDistance(
                    trail.distance,
                )}</span
            >
            <span
                ><i class="fa fa-up-down mr-2"></i>{formatElevation(
                    trail.elevation_gain,
                )}</span
            >
            <span
                ><i class="fa fa-clock mr-2"></i>{formatTimeHHMM(
                    trail.duration,
                )}</span
            >
        </div>
    </div>
</div>

<style>
    .trail-card img {
        object-fit: cover;
        transition: 0.25s ease;
    }

    .trail-card:hover img {
        scale: 1.075;
    }
</style>
