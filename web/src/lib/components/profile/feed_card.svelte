<script lang="ts">
    import type { FeedItem } from "$lib/models/feed";
    import type { Trail } from "$lib/models/trail";
    import { getFileURL, isVideoURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatHTMLAsText,
        formatTimeHHMM,
        formatTimeSince,
    } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    interface Props {
        feedItem: FeedItem;
    }

    let { feedItem }: Props = $props();

    let fullDescription = $state(false)

    const timeSince = formatTimeSince(new Date(feedItem.created ?? ""));

    const photos = (feedItem.expand.item as Trail).photos
</script>

<div class="feed-card p-6 rounded-xl border border-input-border">
    <div class="feed-card-header flex gap-x-4 items-start">
        <img
            class="rounded-full w-10 aspect-square overflow-hidden shrink-0"
            src={feedItem.expand.author?.icon ||
                `https://api.dicebear.com/7.x/initials/svg?seed=${feedItem.expand.author?.preferred_username}&backgroundType=gradientLinear`}
            alt="avatar"
        />
        <div>
            <span class="font-semibold"
                >{feedItem.expand.author?.preferred_username}</span
            >
            <p class="text-sm text-gray-500 mb-3">
                {feedItem.expand.author?.preferred_username}@{feedItem.expand
                    .author?.domain}
            </p>
        </div>
        <div class="basis-full"></div>
        <p class="text-xs text-gray-500 shrink-0">
            {$_(`n-${timeSince.unit}-ago`, {
                values: { n: timeSince.value },
            })}
        </p>
    </div>
    <div class="feed-card-body">
        <h3 class="text-2xl font-semibold !mt-2">
            {feedItem.expand.item.name}
        </h3>
        <div class="flex flex-wrap mt-1 gap-x-4 gap-y-2 text-sm text-gray-500">
            <span
                ><i class="fa fa-left-right mr-2"></i>{formatDistance(
                    feedItem.expand.item.distance,
                )}</span
            >
            <span
                ><i class="fa fa-clock mr-2"></i>{formatTimeHHMM(
                    feedItem.expand.item.duration,
                )}</span
            >
            <span
                ><i class="fa fa-arrow-trend-up mr-2"></i>{formatElevation(
                    feedItem.expand.item.elevation_gain,
                )}</span
            >
            <span
                ><i class="fa fa-arrow-trend-down mr-2"></i>{formatElevation(
                    feedItem.expand.item.elevation_loss,
                )}</span
            >
        </div>
        {#if photos?.length}
            <div
                class="grid gap-[1px] {photos.length > 1
                    ? 'grid-cols-[8fr_5fr]'
                    : 'grid-cols-1'} mt-6"
            >
                {#each photos.slice(0, 3) as photo, i}
                    {#if isVideoURL(photo)}
                        <!-- svelte-ignore a11y_media_has_caption -->
                        <video
                            class="object-cover h-full max-h-80 w-full"
                            autoplay
                            loop
                            src={getFileURL(
                                {
                                    collectionId: "trails",
                                    id: feedItem.item,
                                },
                                photo,
                            )}
                        ></video>
                    {:else}
                        <img
                            class="object-cover h-full max-h-80 w-full"
                            class:row-span-2={i == 0 &&
                                photos.length > 2}
                            src={getFileURL(
                                {
                                    collectionId: "trails",
                                    id: feedItem.item,
                                },
                                photo,
                            )}
                            alt=""
                        />
                    {/if}
                {/each}
            </div>
        {/if}
        {#if feedItem.expand.item.description?.length}
            <p class="text-sm whitespace-pre-wrap mt-6">
                {formatHTMLAsText(
                    !fullDescription
                        ? feedItem.expand.item.description?.substring(0, 100)
                        : feedItem.expand.item.description,
                )}
                {#if (feedItem.expand.item.description?.length ?? 0) > 100 && !fullDescription}
                    <button
                        onclick={(e) => {
                            e.stopPropagation();
                            e.preventDefault();
                            fullDescription = true;
                        }}
                    >
                        ... <span class="underline">{$_("read-more")}</span
                        ></button
                    >
                {/if}
            </p>
        {/if}
    </div>
</div>
