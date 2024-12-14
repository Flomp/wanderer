<script lang="ts">
    import { goto } from "$app/navigation";
    import Tabs from "$lib/components/base/tabs.svelte";
    import TrailDropdown from "$lib/components/trail/trail_dropdown.svelte";
    import WaypointCard from "$lib/components/waypoint/waypoint_card.svelte";
    import { Comment } from "$lib/models/comment";
    import type { Trail } from "$lib/models/trail";

    import {
        comments,
        comments_create,
        comments_delete,
        comments_index,
        comments_update,
    } from "$lib/stores/comment_store";
    import { currentUser } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import {
        formatDistance,
        formatElevation,
        formatTimeHHMM,
    } from "$lib/util/format_util";

    import "photoswipe/style.css";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import Button from "../base/button.svelte";
    import Textarea from "../base/textarea.svelte";
    import CommentCard from "../comment/comment_card.svelte";
    import PhotoGallery from "../photo_gallery.svelte";
    import ShareInfo from "../share_info.svelte";
    import SummitLogTable from "../summit_log/summit_log_table.svelte";
    import { pb } from "$lib/pocketbase";
    import * as M from "maplibre-gl";
    import MapWithElevationMaplibre from "./map_with_elevation_maplibre.svelte";

    export let trail: Trail;
    export let mode: "overview" | "map" | "list" = "map";
    export let markers: M.Marker[] = [];
    export let activeTab = 0;

    const tabs = [
        $_("description"),
        $_("waypoints", { values: { n: 2 } }),
        $_("photos"),
        $_("summit-book"),
        ...($currentUser ? [$_("comment", { values: { n: 2 } })] : []),
    ];

    const trailIsShared =
        (trail.expand?.trail_share_via_trail?.length ?? 0) > 0;


    let thumbnail = trail.photos.length
        ? getFileURL(trail, trail.photos[trail.thumbnail])
        : "/imgs/default_thumbnail.webp";

    let openGallery: (idx?: number) => void;

    let newComment: Comment = new Comment("", 0, "", trail.id ?? "");
    let commentCreateLoading: boolean = false;
    let commentDeleteLoading: boolean = false;

    $: if (activeTab == 4) {
        fetchComments();
    }

    onMount(async () => {});

    function openMarkerPopup(i: number) {
        if((markers[i] as M.Marker).getPopup().isOpen()) {
            return;
        }
        (markers[i] as M.Marker).togglePopup();
    }

    function closeMarkerPopup(i: number) {
        if(!(markers[i] as M.Marker).getPopup().isOpen()) {
            return;
        }
        (markers[i] as M.Marker).togglePopup();
    }

    async function toggleMapFullScreen() {
        goto(`/map/trail/${trail.id!}`);
    }

    async function fetchComments() {
        await comments_index(trail);
    }

    async function createComment() {
        if (!$currentUser || !trail.id) {
            return;
        }
        commentCreateLoading = true;
        newComment.author = $currentUser.id;
        newComment.trail = trail.id;

        comments_create(newComment).then((c) => {
            newComment.text = "";
            c.expand = {
                author: $currentUser!,
            };

            const newCommentList = [c, ...$comments];
            comments.set(newCommentList);
        });

        commentCreateLoading = false;
    }

    async function editComment(data: { comment: Comment; text: string }) {
        data.comment.text = data.text;
        await comments_update(data.comment);
    }

    async function deleteComment(comment: Comment) {
        commentDeleteLoading = true;
        await comments_delete(comment);
        const newCommentList = $comments.filter((c) => c.id !== comment.id);
        comments.set(newCommentList);
        commentDeleteLoading = false;
    }
</script>

<div
    class="trail-info-panel mx-auto {mode == 'list'
        ? ''
        : 'border border-input-border rounded-3xl'} h-full"
    style="max-width: min(100%, 76rem);"
>
    <div class="trail-info-panel-header">
        <section class="relative h-80">
            <img
                class="w-full h-80"
                class:rounded-t-3xl={mode !== "list"}
                src={thumbnail}
                alt=""
            />
            <div
                class="absolute bottom-0 w-full h-2/3 bg-gradient-to-b from-transparent to-black opacity-50"
            ></div>
            {#if (trail.public || trailIsShared) && pb.authStore.model}
                <div
                    class="flex absolute top-8 right-6 {trail.public &&
                    trailIsShared
                        ? 'w-16'
                        : 'w-8'} h-8 rounded-full items-center justify-center bg-white text-primary"
                >
                    {#if trail.public && pb.authStore.model}
                        <span
                            class="tooltip text-2xl"
                            class:mr-3={trail.public && trailIsShared}
                            data-title={$_("public")}
                        >
                            <i class="fa fa-globe"></i>
                        </span>
                    {/if}
                    {#if trailIsShared}
                        <ShareInfo type="trail" subject={trail} large={true}
                        ></ShareInfo>
                    {/if}
                </div>
            {/if}
            <div
                class="flex absolute justify-between items-end w-full bottom-8 left-0 px-8 gap-y-4"
            >
                <div class="text-white">
                    <h4 class="text-4xl font-bold">
                        {trail.name}
                    </h4>
                    {#if trail.date}
                        <h5 class="text-sm">
                            {new Date(trail.date).toLocaleDateString(
                                undefined,
                                {
                                    month: "long",
                                    day: "2-digit",
                                    year: "numeric",
                                    timeZone: "UTC",
                                },
                            )}
                        </h5>
                    {/if}
                    {#if trail.expand.author}
                        <p class="my-3">
                            {$_("by")}
                            <img
                                class="rounded-full w-8 aspect-square mx-1 inline"
                                src={getFileURL(
                                    trail.expand.author,
                                    trail.expand.author.avatar,
                                ) ||
                                    `https://api.dicebear.com/7.x/initials/svg?seed=${trail.expand.author.username}&backgroundType=gradientLinear`}
                                alt="avatar"
                            />
                            <a
                                class="underline"
                                href="/profile/{trail.expand.author.id}"
                                >{trail.expand.author.username}</a
                            >
                        </p>
                    {/if}
                    <div class="flex flex-wrap gap-x-8 gap-y-2 mt-4 mr-8">
                        {#if trail.location}
                            <h3 class="text-lg">
                                <i class="fa fa-location-dot mr-2"></i>
                                {trail.location}
                            </h3>
                        {/if}
                        <h3 class="text-lg">
                            <i class="fa fa-gauge mr-2"></i>
                            {$_(trail.difficulty ?? "?")}
                        </h3>
                    </div>
                </div>
                {#if ($currentUser && $currentUser.id == trail.author) || trail.expand.trail_share_via_trail?.length || trail.public}
                    <TrailDropdown {trail} {mode}></TrailDropdown>
                {/if}
            </div>
        </section>
        {#if mode == "overview"}
            <section
                class="grid grid-cols-2 sm:grid-cols-5 gap-y-4 justify-around py-8 border-b border-input-border"
            >
                <div class="flex flex-col items-center">
                    <span>{$_("distance")}</span>
                    <span class="font-semibold text-lg"
                        >{formatDistance(trail.distance)}</span
                    >
                </div>
                <div class="flex flex-col items-center">
                    <span>{$_("est-duration")}</span>
                    <span class="font-semibold text-lg"
                        >{formatTimeHHMM(trail.duration)}</span
                    >
                </div>
                <div class="flex flex-col items-center">
                    <span>{$_("elevation-gain")}</span>
                    <span class="font-semibold text-lg"
                        >{formatElevation(trail.elevation_gain)}</span
                    >
                </div>
                <div class="flex flex-col items-center">
                    <span>{$_("elevation-loss")}</span>
                    <span class="font-semibold text-lg"
                        >{formatElevation(trail.elevation_loss)}</span
                    >
                </div>
                {#if trail.expand.category}
                    <div class="flex flex-col items-center">
                        <span>{$_("category")}</span>
                        <span class="font-semibold text-lg"
                            >{$_(trail.expand.category.name)}</span
                        >
                    </div>
                {/if}
            </section>
        {/if}
        {#if trail.tags && trail.tags.length > 0}
            <hr class="border-separator" />
            <section class="flex p-8 gap-4 text-gray-600">
                {#each trail.tags as tag}
                    <span class="py-2 px-4 border rounded-full">{tag}</span>
                {/each}
            </section>
            <hr class="border-separator" />
        {/if}
    </div>
    <section class="trail-info-panel-tabs px-4 py-2 bg-background sticky top-0">
        <Tabs {tabs} bind:activeTab></Tabs>
    </section>
    <section class="trail-info-panel-content px-8">
        <div
            class="grid grid-cols-1 my-4 gap-8"
            class:md:grid-cols-[1fr_18rem]={mode == "overview"}
        >
            {#if activeTab == 0}
                <article class="text-justify whitespace-pre-line text-sm">
                    {trail.description}
                </article>
            {/if}
            {#if activeTab == 1}
                <ul>
                    {#each trail.expand.waypoints ?? [] as waypoint, i}
                        <li
                            on:mouseenter={() => openMarkerPopup(i)}
                            on:mouseleave={() => closeMarkerPopup(i)}
                        >
                            <WaypointCard {waypoint}></WaypointCard>
                        </li>
                    {/each}
                </ul>
            {/if}
            {#if activeTab == 2}
                <div
                    id="photo-gallery"
                    class="grid grid-cols-1 {mode == 'overview'
                        ? 'sm:grid-cols-2 md:grid-cols-3'
                        : ''} gap-4"
                >
                    <PhotoGallery
                        photos={trail.photos.map((p) => getFileURL(trail, p))}
                        bind:open={openGallery}
                    ></PhotoGallery>
                    {#each trail.photos ?? [] as photo, i}
                        <!-- svelte-ignore a11y-click-events-have-key-events -->
                        <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
                        <img
                            class="rounded-xl cursor-pointer hover:scale-105 transition-transform"
                            on:click={() => openGallery(i)}
                            src={getFileURL(trail, photo)}
                            alt=""
                        />
                    {/each}
                </div>
            {/if}
            {#if activeTab == 3}
                <div class="overflow-x-auto">
                    <SummitLogTable
                        summitLogs={trail.expand.summit_logs}
                        showAuthor
                        showRoute
                        showPhotos
                    ></SummitLogTable>
                </div>
            {/if}
            {#if activeTab == 4}
                <div>
                    {#if $currentUser}
                        <div class="flex items-center gap-4">
                            <img
                                class="rounded-full w-10 aspect-square"
                                src={getFileURL(
                                    $currentUser,
                                    $currentUser.avatar,
                                ) ||
                                    `https://api.dicebear.com/7.x/initials/svg?seed=${$currentUser.username}&backgroundType=gradientLinear`}
                                alt="avatar"
                            />
                            <div class="basis-full">
                                <Textarea
                                    bind:value={newComment.text}
                                    rows={2}
                                    placeholder="Add comment..."
                                ></Textarea>
                            </div>
                        </div>
                        <div class="flex justify-end mt-3">
                            <Button
                                on:click={createComment}
                                loading={commentCreateLoading}
                                secondary={true}
                                disabled={newComment.text.length == 0}
                                >Comment</Button
                            >
                        </div>
                    {/if}
                    <ul class="space-y-4">
                        {#each $comments ?? [] as comment}
                            <li>
                                <CommentCard
                                    {comment}
                                    mode={comment.author == $currentUser?.id
                                        ? "edit"
                                        : "show"}
                                    on:delete={(e) => deleteComment(e.detail)}
                                    on:edit={(e) => editComment(e.detail)}
                                ></CommentCard>
                            </li>
                        {/each}
                    </ul>
                </div>
            {/if}
            {#if mode == "overview"}
                <div class="relative h-72 rounded-xl overflow-hidden">
                    <MapWithElevationMaplibre
                        trails={[trail]}
                        showElevation={false}
                        showStyleSwitcher={false}
                        showFullscreen={true}
                        mapOptions={{ attributionControl: false }}
                        on:fullscreen={toggleMapFullScreen}
                        bind:markers
                    ></MapWithElevationMaplibre>
                </div>
            {/if}
        </div>
    </section>
</div>

<style>
    .trail-info-panel img {
        object-fit: cover;
    }
</style>
