<script lang="ts">
    import { page } from "$app/stores";
    import { follows_index } from "$lib/stores/follow_store.js";
    import { getFileURL } from "$lib/util/file_util.js";
    import { _ } from "svelte-i18n";
    export let data;

    let loading: boolean = false;

    $: pagination = {
        page: data.follows.page,
        totalPages: data.follows.totalPages,
    };

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
        data.follows = await follows_index(
            { followee: $page.params.id },
            pagination.page,
        );
    }
</script>

<svelte:window on:scroll={onListScroll} />
<div>
    <h1 class="text-3xl font-semibold mb-4">{$_($page.params.type)}</h1>
    <ul class="space-y-4">
        {#each data.follows.items as follow}
            {#if !follow.expand?.follower.private}
                <a href="/profile/{follow.expand?.follower.id}">
                    <li
                        class="flex items-center gap-x-4 hover:bg-menu-item-background-hover p-4"
                    >
                        <img
                            class="rounded-full w-10 aspect-square overflow-hidden"
                            src={getFileURL(
                                follow.expand?.follower ?? {},
                                follow.expand?.follower.avatar,
                            ) ||
                                `https://api.dicebear.com/7.x/initials/svg?seed=${data.user.username}&backgroundType=gradientLinear`}
                            alt="avatar"
                        />
                        <p class="text-lg font-medium">
                            {follow.expand?.follower.username}
                        </p>
                    </li>
                </a>
            {:else}
                <li
                    class="flex items-center gap-x-4 hover:bg-menu-item-background-hover p-4"
                >
                    <img
                        class="rounded-full w-10 aspect-square overflow-hidden"
                        src={getFileURL(
                            follow.expand?.follower ?? {},
                            follow.expand?.follower.avatar,
                        ) ||
                            `https://api.dicebear.com/7.x/initials/svg?seed=${data.user.username}&backgroundType=gradientLinear`}
                        alt="avatar"
                    />
                    <p class="text-lg font-medium">
                        {follow.expand?.follower.username}
                    </p>
                </li>
            {/if}
        {/each}
    </ul>
</div>
