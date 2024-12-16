<script lang="ts">
    import type { List } from "$lib/models/list";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import ShareInfo from "../share_info.svelte";
    export let list: List;
    export let active: boolean = false;

    $: cumulativeDistance = list.expand?.trails?.reduce(
        (s, b) => s + b.distance!,
        0,
    );

    $: cumulativeElevationGain = list.expand?.trails?.reduce(
        (s, b) => s + b.elevation_gain!,
        0,
    );

    $: cumulativeElevationLoss = list.expand?.trails?.reduce(
        (s, b) => s + b.elevation_loss!,
        0,
    );

    $: cumulativeDuration = list.expand?.trails?.reduce(
        (s, b) => s + b.duration!,
        0,
    );

    $: listIsShared = (list.expand?.list_share_via_list?.length ?? 0) > 0;
</script>

<div
    class="flex items-start gap-6 p-4 hover:bg-menu-item-background-hover rounded-xl transition-colors cursor-pointer"
    class:bg-menu-item-background-hover={active}
>
    <img
        class="w-16 md:w-20 aspect-square rounded-full object-cover"
        src={list.avatar
            ? getFileURL(list, list.avatar)
            : "/imgs/default_list_thumbnail.webp"}
        alt="avatar"
    />

    <div class="self-start min-w-0 basis-full transition-transform">
        <div class="flex items-center gap-3">
            <h5 class="text-xl font-semibold overflow-hidden overflow-ellipsis">
                {list.name}
            </h5>
            {#if list.public}
                <span class="tooltip" data-title={$_("public")}>
                    <i class="fa fa-globe"></i>
                </span>
            {/if}
            {#if listIsShared}
                <ShareInfo type="list" subject={list}></ShareInfo>
            {/if}
        </div>
        {#if list.expand?.author}
            <p class="text-xs text-gray-500 my-2">
                {$_("by")}
                <img
                    class="rounded-full w-5 aspect-square mx-1 inline"
                    src={getFileURL(
                        list.expand.author,
                        list.expand.author.avatar,
                    ) ||
                        `https://api.dicebear.com/7.x/initials/svg?seed=${list.expand.author.username}&backgroundType=gradientLinear`}
                    alt="avatar"
                />
                {list.expand?.author.username}
            </p>
        {/if}
        <div
            class="grid grid-cols-2 mt-1 mb-2 gap-x-4 gap-y-1 text-sm text-gray-500 whitespace-nowrap flex-wrap"
        >
            <span
                ><i class="fa fa-left-right mr-2"></i>{formatDistance(
                    cumulativeDistance,
                )}</span
            >
            <span
                ><i class="fa fa-clock mr-2"></i>{formatTimeHHMM(
                    cumulativeDuration,
                )}</span
            >
            <span
                ><i class="fa fa-arrow-trend-up mr-2"></i>{formatElevation(
                    cumulativeElevationGain,
                )}</span
            >
            <span
                ><i class="fa fa-arrow-trend-down mr-2"></i>{formatElevation(
                    cumulativeElevationLoss,
                )}</span
            >
        </div>
        <p class="text-sm text-gray-500 mb-2">
            {list.expand?.trails?.length ?? 0}
            {$_("trail", {
                values: { n: list.expand?.trails?.length ?? 0 },
            })}
        </p>
        <p
            class="text-gray-500 text-sm mr-8 whitespace-pre-wrap {active
                ? ''
                : 'max-h-24 overflow-hidden text-ellipsis'}"
        >
            {!active ? list.description?.substring(0, 100) : list.description}
            {#if (list.description?.length ?? 0) > 100 && !active}
                ...
            {/if}
        </p>
    </div>
</div>
