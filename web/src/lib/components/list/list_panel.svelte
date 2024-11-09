<script lang="ts">
    import type { List } from "$lib/models/list";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    import TrailListItem from "../trail/trail_list_item.svelte";
    import type { Trail } from "$lib/models/trail";
    import { createEventDispatcher } from "svelte";
    import { currentUser } from "$lib/stores/user_store";
    import Dropdown from "../base/dropdown.svelte";
    import ShareInfo from "../share_info.svelte";

    export let list: List;

    const dispatch = createEventDispatcher();

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

    $: allowEdit =
        list.author == $currentUser?.id ||
        list.expand?.list_share_via_list?.some((s) => s.permission == "edit");

    $: dropdownItems = [
        ...(list.author == $currentUser?.id
            ? [{ text: $_("share"), value: "share", icon: "share" }]
            : []),
        ...(allowEdit
            ? [{ text: $_("edit"), value: "edit", icon: "pen" }]
            : []),
        ...(list.author == $currentUser?.id
            ? [{ text: $_("delete"), value: "delete", icon: "trash" }]
            : []),
    ];

    $: listIsShared = (list.expand?.list_share_via_list?.length ?? 0) > 0;

    let fullDescription: boolean = false;

    function handleTrailSelect(trail: Trail, index: number) {
        dispatch("click", { trail, index });
    }

    function handleTrailMouseEnter(trail: Trail, index: number) {
        dispatch("mouseenter", { trail, index });
    }

    function handleTrailMouseLeave(trail: Trail, index: number) {
        dispatch("mouseleave", { trail, index });
    }
</script>

<div class="relative">
    {#if listIsShared}
        <div
            class="absolute top-8 right-8 bg-white rounded-full w-8 py-1 text-center"
        >
            <ShareInfo type="list" subject={list}></ShareInfo>
        </div>
    {/if}

    {#if dropdownItems.length}
        <div class="absolute bottom-8 right-8">
            <Dropdown
                items={dropdownItems}
                on:change
                let:toggleMenu={openDropdown}
                ><button
                    class="rounded-full bg-white text-black hover:bg-gray-200 focus:ring-4 ring-gray-100/50 transition-colors h-12 w-12"
                    on:click={openDropdown}
                >
                    <i class="fa fa-ellipsis-vertical"></i>
                </button></Dropdown
            >
        </div>
    {/if}
    {#if list.avatar}
        <img
            class="w-full object-cover"
            src={getFileURL(list, list.avatar)}
            alt="avatar"
        />
    {:else}
        <div class="flex w-full shrink-0 items-center justify-center min-h-72">
            <i class="fa fa-table-list text-5xl"></i>
        </div>
    {/if}
</div>

<div class="p-4 md:p-6">
    <h4 class="text-2xl font-semibold mb-4">{list.name}</h4>

    <hr />
    <div
        class="grid grid-cols-2 my-4 gap-4 font-semibold whitespace-nowrap flex-wrap justify-around"
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
    <hr class="mb-4" />
    <p
        class="text-gray-500 whitespace-pre-wrap {fullDescription
            ? ''
            : 'max-h-24 overflow-hidden text-ellipsis'}"
    >
        {!fullDescription
            ? list.description?.substring(0, 100)
            : list.description}
        {#if (list.description?.length ?? 0) > 100 && !fullDescription}
            <button on:click={() => (fullDescription = true)}>
                ... <span class="underline">{$_("read-more")}</span></button
            >
        {/if}
    </p>
    <h5 class="text-xl font-semibold my-4">
        {list.trails?.length ?? 0}
        {$_("trail", { values: { n: list.trails?.length ?? 0 } })}
    </h5>
    <div class="space-y-2">
        {#each list.expand?.trails ?? [] as trail, i}
            <div
                role="presentation"
                on:click={() => handleTrailSelect(trail, i)}
                on:mouseenter={() => handleTrailMouseEnter(trail, i)}
                on:mouseleave={() => handleTrailMouseLeave(trail, i)}
            >
                <TrailListItem {trail} showDescription={false}></TrailListItem>
            </div>
        {/each}
    </div>
</div>
