<script lang="ts">
    import type { Activity } from "$lib/models/activitypub/activity";
    import type { Actor } from "$lib/models/activitypub/actor";
    import { isVideoURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";
    import { _ } from "svelte-i18n";
    interface Props {
        activity: Activity;
        user: Actor;
    }

    let { activity, user }: Props = $props();

    let fullDescription: boolean = $state(false);

    if(!(activity.object.attachment instanceof Array)) {
        activity.object.attachment = [activity.object.attachment]
    }
    const photos = activity.object.attachment?.filter(
        (a: any) => a.mediaType.startsWith("image") ?? [],
    );    
</script>

<div class="activity-card p-6 rounded-xl border border-input-border">
    <div class="flex gap-x-4 items-start">
        <img
            class="rounded-full w-10 aspect-square overflow-hidden"
            src={user.icon ||
                `https://api.dicebear.com/7.x/initials/svg?seed=${user.username}&backgroundType=gradientLinear`}
            alt="avatar"
        />
        <div>
            <a class="underline" href="/profile/@{user.username}">{user.username}</a>

            {activity.object.type === "Trail"
                ? $_("planned-a-trail")
                : $_("completed-a-trail")}
            <p class="text-xs text-gray-500 mb-3">
                {new Date(activity.published).toLocaleDateString(undefined, {
                    month: "long",
                    day: "2-digit",
                    year: "numeric",
                    timeZone: "UTC",
                })}
            </p>
        </div>
    </div>
    <h3 class="text-2xl font-semibold !mt-2">{activity.object.name}</h3>
    <div class="flex flex-wrap mt-1 mb-4 gap-x-4 gap-y-2 text-sm text-gray-500">
        <span
            ><i class="fa fa-left-right mr-2"></i>{formatDistance(
                activity.object.distance,
            )}</span
        >
        <span
            ><i class="fa fa-clock mr-2"></i>{formatTimeHHMM(
                activity.object.duration,
            )}</span
        >
        <span
            ><i class="fa fa-arrow-trend-up mr-2"></i>{formatElevation(
                activity.object.elevation_gain,
            )}</span
        >
        <span
            ><i class="fa fa-arrow-trend-down mr-2"></i>{formatElevation(
                activity.object.elevation_loss,
            )}</span
        >
    </div>
    {#if photos.length}
        <div
            class="grid gap-[1px] {photos.length > 1
                ? 'grid-cols-[8fr_5fr]'
                : 'grid-cols-1'}"
        >
            {#each photos.slice(0, 3) as photo, i}
                {#if isVideoURL(photo.url)}
                    <!-- svelte-ignore a11y_media_has_caption -->
                    <video
                        class="object-cover h-full max-h-80 w-full"
                        autoplay
                        loop
                        src={photo.url}
                    ></video>
                {:else}
                    <img
                        class="object-cover h-full max-h-80 w-full"
                        class:row-span-2={i == 0 && photos.length > 2}
                        src={photo.url}
                        alt=""
                    />
                {/if}
            {/each}
        </div>
    {/if}
    {#if activity.object.content?.length}
        <p class="text-sm whitespace-pre-wrap mt-4">
            {!fullDescription
                ? activity.object.content.substring(0, 100)
                : activity.object.content}
            {#if (activity.object.content.length ?? 0) > 100 && !fullDescription}
                <button
                    onclick={(e) => {
                        e.stopPropagation();
                        e.preventDefault();
                        fullDescription = true;
                    }}
                >
                    ... <span class="underline">{$_("read-more")}</span></button
                >
            {/if}
        </p>
    {/if}
</div>
