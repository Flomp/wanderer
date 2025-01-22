<script lang="ts">
    import ActivityCard from "$lib/components/profile/activity_card.svelte";
    import { activities_index } from "$lib/stores/activity_store.js";
    import { getFileURL } from "$lib/util/file_util.js";
    import { _ } from "svelte-i18n";
    import emptyStateTrailDark from "$lib/assets/svgs/empty_states/empty_state_trail_dark.svg";
    import emptyStateTrailLight from "$lib/assets/svgs/empty_states/empty_state_trail_light.svg";
    import { theme } from "$lib/stores/theme_store.js";

    let { data = $bindable() } = $props();

    let loading: boolean = false;

    let pagination = $derived({
        page: data.activities.page,
        totalPages: data.activities.totalPages,
    });

    async function onListScroll(e: Event) {
        if (
            window.innerHeight + window.scrollY >=
                0.8 * document.body.offsetHeight &&
            pagination.page !== pagination.totalPages &&
            !loading
        ) {
            loading = true;
            await loadNextPage();
            loading = false;
        }
    }

    async function loadNextPage() {
        pagination.page += 1;
        data.activities = await activities_index(data.user.id, pagination.page);
    }
</script>

<svelte:window onscroll={onListScroll} />

<svelte:head>
    <title>{$_("profile")} | wanderer</title>
</svelte:head>

<div class="space-y-6">
    <div class="space-y-4">
        <h4 class="text-xl font-semibold">
            {$_("about")}
            {data.user.username}
            {#if data.isOwnProfile && data.settings.bio?.length}
                <a class="ml-4" href="/settings/profile"
                    ><i class="fa fa-pen text-base"></i></a
                >
            {/if}
        </h4>
        {#if data.settings.bio?.length}
            <p class="whitespace-pre-wrap text-sm">{data.settings.bio}</p>
        {:else if data.isOwnProfile}
            <a class="btn-primary inline-block" href="/settings/profile"
                >+ {$_("add-bio")}</a
            >
        {:else}
            <p class="w-full text-center text-gray-500 text-sm">
                {$_("empty-bio", { values: { username: data.user.username } })}
            </p>
        {/if}
    </div>
    <div class="space-y-4">
        <h4 class="text-xl font-semibold">
            {$_("list", { values: { n: 2 } })}
        </h4>
        <div class="flex gap-x-4 overflow-x-scroll">
            {#if !data.lists.length && data.isOwnProfile}
                <a class="btn-primary inline-block" href="/lists/edit/new"
                    >+ {$_("new-list")}</a
                >
            {:else if !data.lists.length}
                <p class="w-full text-center text-gray-500 text-sm">
                    {$_("empty-lists", {
                        values: { username: data.user.username },
                    })}
                </p>
            {/if}
            {#each data.lists as list}
                <a
                    href="/lists?list={list.id}"
                    class="relative w-64 h-48 rounded-xl overflow-hidden group shrink-0"
                >
                    <img
                        class="w-full h-full object-cover transition-transform group-hover:scale-110"
                        src={list.avatar
                            ? getFileURL(list, list.avatar)
                            : $theme === "light"
                              ? emptyStateTrailLight
                              : emptyStateTrailDark}
                        alt="avatar"
                    />

                    <div
                        class="absolute bottom-0 w-full h-2/3 bg-gradient-to-b from-transparent to-black opacity-50"
                    ></div>
                    <h5
                        class="absolute bottom-4 left-4 font-semibold text-white"
                    >
                        {list.name}
                    </h5>
                </a>
            {/each}
        </div>
    </div>
    <div class="space-y-4">
        <h4 class="text-xl font-semibold">Timeline</h4>
        {#if !data.activities.items.length && data.isOwnProfile}
            <a class="btn-primary inline-block" href="/trails/edit/new"
                >+ {$_("new-trail")}</a
            >
        {:else if !data.activities.items.length}
            <p class="w-full text-center text-gray-500 text-sm">
                {$_("empty-activities", {
                    values: { username: data.user.username },
                })}
            </p>
        {/if}
        {#each data.activities.items as activity}
            <a
                class="block"
                href={activity.type == "trail"
                    ? `/trail/view/${activity.id}`
                    : `/trail/view/${activity.trail_id}?t=3`}
            >
                <ActivityCard {activity} user={data.user}></ActivityCard>
            </a>
        {/each}
    </div>
</div>
