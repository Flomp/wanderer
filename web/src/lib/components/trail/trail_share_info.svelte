<script lang="ts">
    import type { Trail } from "$lib/models/trail";
    import type { User } from "$lib/models/user";
    import { pb } from "$lib/pocketbase";
    import { users_show } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import { _ } from "svelte-i18n";
    import { fly, slide } from "svelte/transition";

    export let trail: Trail;
    export let large: boolean = false;

    let showInfo: boolean = false;

    let loading: boolean = false;

    let infoLoaded: boolean = false;

    let trailIsOwned: boolean = trail.author == pb.authStore.model?.id;

    let author: User;

    async function fetchInfo() {
        if (!infoLoaded) {
            loading = true;
            loading = false;
            if (trailIsOwned) {
                for (const share of trail.expand.trail_share_via_trail ?? []) {
                    share.expand = {
                        user: await users_show(share.user),
                    };
                }
            } else {
                author = await users_show(trail.author!);
            }

            infoLoaded = true;
        }

        showInfo = true;
    }
</script>

<div class="relative trail-share-info">
    <i
        class="fa fa-share-nodes"
        class:text-xl={large}
        role="button"
        tabindex="0"
        on:mouseenter={fetchInfo}
        on:mouseleave={() => (showInfo = false)}
    ></i>
    {#if showInfo}
        <ul
            class="menu absolute z-10 top-7 bg-menu-background border border-input-border rounded-xl shadow-md overflow-hidden p-3 space-y-3 text-content"
            in:fly={{y: -10, duration: 150}} out:fly={{y: -10, duration: 150}}
            >
            {#if loading}
                <div class="spinner spinner-dark"></div>
            {:else if infoLoaded && trail.expand.trail_share_via_trail}
                {#if trailIsOwned}
                    {#each trail.expand.trail_share_via_trail as share}
                        {#if share.expand}
                            <li>
                                <div class="flex items-center mr-8">
                                    <img
                                        class="rounded-full w-8 aspect-square mr-2"
                                        src={getFileURL(
                                            share.expand.user,
                                            share.expand.user.avatar,
                                        ) ||
                                            `https://api.dicebear.com/7.x/initials/svg?seed=${share.expand.user.username}&backgroundType=gradientLinear`}
                                        alt="avatar"
                                    />
                                    <p class="font-semibold text-base">
                                        {share.expand.user.username}
                                    </p>
                                    <span class="mx-2 text-sm text-gray-500"
                                        >{$_("can")}</span
                                    >
                                    <p class="whitespace-nowrap text-base">
                                        {#if share.permission == "view"}
                                            <i class="fa fa-eye"></i>
                                        {:else}
                                            <i class="fa fa-pen"></i>
                                        {/if}
                                    </p>
                                </div>
                            </li>
                        {/if}
                    {/each}
                {:else if author}
                    <li>
                        <p class="text-xs text-gray-500 mb-2">
                            {$_("shared-by")}
                        </p>
                        <div class="flex items-center mr-8">
                            <img
                                class="rounded-full w-8 aspect-square mr-2"
                                src={getFileURL(author, author.avatar) ||
                                    `https://api.dicebear.com/7.x/initials/svg?seed=${author.username}&backgroundType=gradientLinear`}
                                alt="avatar"
                            />
                            <p class="font-semibold">
                                {author.username}
                            </p>
                        </div>
                        <p class="text-xs text-gray-500 mt-2">
                            {$_("you-can")}
                        </p>
                        <p class="whitespace-nowrap">
                            {#if trail.expand.trail_share_via_trail[0].permission == "view"}
                                <i class="fa fa-eye mr-1"></i>
                                {$_("view")}
                            {:else}
                                <i class="fa fa-pen mr-1"></i>
                                {$_("edit")}
                            {/if}
                        </p>
                    </li>
                {/if}
            {/if}
        </ul>
    {/if}
</div>
