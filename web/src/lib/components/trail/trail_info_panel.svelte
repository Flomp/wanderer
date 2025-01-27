<script lang="ts">
    import { goto } from "$app/navigation";
    import Tabs from "$lib/components/base/tabs.svelte";
    import TrailDropdown from "$lib/components/trail/trail_dropdown.svelte";
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

    import { browser } from "$app/environment";
    import emptyStateTrailDark from "$lib/assets/svgs/empty_states/empty_state_trail_dark.svg";
    import emptyStateTrailLight from "$lib/assets/svgs/empty_states/empty_state_trail_light.svg";
    import { pb } from "$lib/pocketbase";
    import { theme } from "$lib/stores/theme_store";
    import { show_toast } from "$lib/stores/toast_store";
    import * as M from "maplibre-gl";
    import "photoswipe/style.css";
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import Button from "../base/button.svelte";
    import SkeletonNotificationCard from "../base/skeleton_notification_card.svelte";
    import Textarea from "../base/textarea.svelte";
    import CommentCard from "../comment/comment_card.svelte";
    import EmptyStateComment from "../empty_states/empty_state_comment.svelte";
    import EmptyStateDescription from "../empty_states/empty_state_description.svelte";
    import EmptyStatePhotos from "../empty_states/empty_state_photos.svelte";
    import PhotoGallery from "../photo_gallery.svelte";
    import ShareInfo from "../share_info.svelte";
    import SummitLogTable from "../summit_log/summit_log_table.svelte";
    import MapWithElevationMaplibre from "./map_with_elevation_maplibre.svelte";
    import TrailTimeline from "./trail_timeline.svelte";

    interface Props {
        trail: Trail;
        mode?: "overview" | "map" | "list";
        markers?: M.Marker[];
        activeTab?: number;
    }

    let { trail, mode = "map", markers = [], activeTab = 0 }: Props = $props();

    const tabs = [
        $_("summit-book"),
        $_("photos"),
        ...(pb.authStore.model ? [$_("comment", { values: { n: 2 } })] : []),
    ];

    const trailIsShared =
        (trail.expand?.trail_share_via_trail?.length ?? 0) > 0;

    let gallery: PhotoGallery;

    let newComment: Comment = $state({
        text: "",
        rating: 0,
        author: "",
        trail: trail.id ?? "",
    });

    let commentsLoading: boolean = $state(activeTab == 4);
    let commentCreateLoading: boolean = $state(false);
    let commentDeleteLoading: boolean = false;

    let fullDescription: boolean = $state(false);

    onMount(async () => {});

    function openMarkerPopup(i: number) {
        if ((markers[i] as M.Marker).getPopup().isOpen()) {
            return;
        }
        (markers[i] as M.Marker).togglePopup();
    }

    function closeMarkerPopup(i: number) {
        if (!(markers[i] as M.Marker).getPopup().isOpen()) {
            return;
        }
        (markers[i] as M.Marker).togglePopup();
    }

    async function toggleMapFullScreen() {
        goto(`/map/trail/${trail.id!}`);
    }

    async function fetchComments() {
        commentsLoading = true;
        await comments_index(trail);
        commentsLoading = false;
    }

    async function createComment() {
        if (!$currentUser || !trail.id) {
            return;
        }
        commentCreateLoading = true;
        newComment.author = $currentUser.id;
        newComment.trail = trail.id;

        try {
            const c = await comments_create(newComment);
            newComment.text = "";
            c.expand = {
                author: { ...$currentUser!, private: false },
            };

            const newCommentList = [c, ...$comments];
            comments.set(newCommentList);
        } catch (e) {
            show_toast({
                icon: "close",
                type: "error",
                text: $_("error-posting-comment"),
            });
        } finally {
            commentCreateLoading = false;
        }
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
    let thumbnail = $derived(
        trail.photos.length
            ? getFileURL(trail, trail.photos[trail.thumbnail ?? 0])
            : $theme === "light"
              ? emptyStateTrailLight
              : emptyStateTrailDark,
    );
    $effect(() => {
        if (browser && activeTab == 4) {
            fetchComments();
        }
    });
</script>

<div
    class="trail-info-panel mx-auto {mode == 'list'
        ? ''
        : 'border border-input-border rounded-3xl'} h-full"
    class:overflow-y-scroll={mode !== "overview"}
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
                class="absolute bottom-0 w-full h-full bg-gradient-to-b from-transparent to-black opacity-60"
            ></div>
            {#if (trail.public || trailIsShared) && pb.authStore.model}
                <div
                    class="flex absolute top-6 right-6 {trail.public &&
                    trailIsShared
                        ? 'w-16'
                        : 'w-8'} h-8 rounded-full items-center justify-center bg-white text-primary"
                >
                    {#if trail.public && pb.authStore.model}
                        <span
                            class="tooltip"
                            class:mr-3={trail.public && trailIsShared}
                            data-title={$_("public")}
                        >
                            <i class="fa fa-globe"></i>
                        </span>
                    {/if}
                    {#if trailIsShared}
                        <ShareInfo type="trail" subject={trail}></ShareInfo>
                    {/if}
                </div>
            {/if}
            {#if mode == "map"}
                <button
                    aria-label="Back"
                    class="btn-icon hover:bg-white hover:text-black ring-input-ring text-white absolute top-6 left-6"
                    onclick={() => history.back()}
                >
                    <i class="fa fa-arrow-left"></i>
                </button>
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
                    {#if trail.expand?.author}
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
                            {#if !trail.expand.author.private}
                                <a
                                    class="underline"
                                    href="/profile/{trail.expand.author.id}"
                                    >{trail.expand.author.username}</a
                                >
                            {:else}
                                <span>{trail.expand.author.username}</span>
                            {/if}
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
                {#if ($currentUser && $currentUser.id == trail.author) || trail.expand?.trail_share_via_trail?.length || trail.public}
                    <TrailDropdown {trail} {mode}></TrailDropdown>
                {/if}
            </div>
        </section>
        <section
            class="grid grid-cols-2 sm:grid-cols-5 gap-y-4 py-4 border-b border-input-border px-3"
        >
            <div class="flex flex-col items-center">
                <span class="font-medium text-center"
                    >{#if mode == "overview"}
                        {$_("distance")}
                    {:else}
                        <i class="fa fa-left-right"></i>
                    {/if}</span
                >
                <span class="">{formatDistance(trail.distance)}</span>
            </div>
            <div class="flex flex-col items-center">
                <span class="font-medium text-center"
                    >{#if mode == "overview"}
                        {$_("est-duration")}
                    {:else}
                        <i class="fa fa-clock"></i>
                    {/if}</span
                >
                <span class="">{formatTimeHHMM(trail.duration)}</span>
            </div>
            <div class="flex flex-col items-center">
                <span class="font-medium text-center"
                    >{#if mode == "overview"}
                        {$_("elevation-gain")}
                    {:else}
                        <i class="fa fa-arrow-trend-up"></i>
                    {/if}</span
                >
                <span class="">{formatElevation(trail.elevation_gain)}</span>
            </div>
            <div class="flex flex-col items-center">
                <span class="font-medium text-center"
                    >{#if mode == "overview"}
                        {$_("elevation-loss")}
                    {:else}
                        <i class="fa fa-arrow-trend-down"></i>
                    {/if}</span
                >
                <span class="">{formatElevation(trail.elevation_loss)}</span>
            </div>
            {#if trail.expand?.category}
                <div class="flex flex-col items-center">
                    <span class="font-medium text-center"
                        >{#if mode == "overview"}
                            {$_("category")}
                        {:else}
                            <i class="fa fa-route"></i>
                        {/if}</span
                    >
                    <span class="">{$_(trail.expand.category.name)}</span>
                </div>
            {/if}
        </section>
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
    <section class="trail-info-panel-content px-8">
        <div
            class="grid grid-cols-1 my-4 gap-8"
            class:xl:grid-cols-[1fr_18rem]={mode == "overview"}
        >
            <div class="order-1 xl:-order-1">
                <h4 class="text-2xl font-semibold my-4">
                    {$_("description")}
                </h4>
                {#if trail.description?.length}
                    <article class="text-justify whitespace-pre-line text-sm">
                        {!fullDescription
                            ? trail.description?.substring(0, 300)
                            : trail.description}
                        {#if (trail.description?.length ?? 0) > 300 && !fullDescription}
                            <button
                                onclick={(e) => {
                                    e.stopPropagation();
                                    e.preventDefault();
                                    fullDescription = true;
                                }}
                            >
                                ... <span class="underline"
                                    >{$_("read-more")}</span
                                ></button
                            >
                        {/if}
                    </article>
                {:else}
                    <EmptyStateDescription></EmptyStateDescription>
                {/if}
                <h4 class="text-2xl font-semibold mb-6 mt-12">{$_("route")}</h4>
                {#if mode === "overview"}
                    <div
                        class="relative border border-input-border rounded-xl p-2 mb-6 text-xs"
                        id="epc-container"
                    ></div>
                {/if}
                <TrailTimeline
                    {trail}
                    onmouseenter={openMarkerPopup}
                    onmouseleave={closeMarkerPopup}
                ></TrailTimeline>

                <div class="mb-6 mt-12">
                    <Tabs {tabs} bind:activeTab></Tabs>
                </div>
                {#if activeTab == 0}
                    <div class="overflow-x-auto">
                        <SummitLogTable
                            summitLogs={trail.expand?.summit_logs}
                            showAuthor
                            showRoute
                            showPhotos
                        ></SummitLogTable>
                    </div>
                {/if}
                {#if activeTab == 1}
                    {#if trail.photos.length}
                        <div
                            id="photo-gallery"
                            class="grid grid-cols-1 {mode == 'overview'
                                ? 'sm:grid-cols-2 md:grid-cols-3'
                                : ''} gap-4"
                        >
                            <PhotoGallery
                                photos={trail.photos.map((p) =>
                                    getFileURL(trail, p),
                                )}
                                bind:this={gallery}
                            ></PhotoGallery>
                            {#each trail.photos ?? [] as photo, i}
                                <!-- svelte-ignore a11y_click_events_have_key_events -->
                                <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
                                <img
                                    class="rounded-xl cursor-pointer hover:scale-105 transition-transform"
                                    onclick={() => gallery.openGallery(i)}
                                    src={getFileURL(trail, photo)}
                                    alt=""
                                />
                            {/each}
                        </div>
                    {:else}
                        <EmptyStatePhotos></EmptyStatePhotos>
                    {/if}
                {/if}
                {#if activeTab == 2}
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
                                    onclick={createComment}
                                    loading={commentCreateLoading}
                                    secondary={true}
                                    disabled={commentCreateLoading ||
                                        newComment.text.length == 0}
                                    >Comment</Button
                                >
                            </div>
                        {/if}
                        {#if commentsLoading}
                            {#each { length: 3 } as _, index}
                                <SkeletonNotificationCard
                                ></SkeletonNotificationCard>
                            {/each}
                        {:else if $comments.length == 0}
                            <div class="my-4">
                                <EmptyStateComment></EmptyStateComment>
                            </div>
                        {:else}
                            <ul class="space-y-4">
                                {#each $comments ?? [] as comment}
                                    <li>
                                        <CommentCard
                                            {comment}
                                            mode={comment.author ==
                                            $currentUser?.id
                                                ? "edit"
                                                : "show"}
                                            ondelete={deleteComment}
                                            onedit={editComment}
                                        ></CommentCard>
                                    </li>
                                {/each}
                            </ul>
                        {/if}
                    </div>
                {/if}
            </div>

            {#if mode == "overview"}
                <div
                    class="block xl:sticky top-4 h-72 rounded-xl overflow-hidden"
                >
                    <MapWithElevationMaplibre
                        trails={[trail]}
                        activeTrail={0}
                        waypoints={trail.expand?.waypoints}
                        showElevation={true}
                        elevationProfileContainer={"epc-container"}
                        showStyleSwitcher={false}
                        showFullscreen={true}
                        mapOptions={{ attributionControl: false }}
                        onfullscreen={toggleMapFullScreen}
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
