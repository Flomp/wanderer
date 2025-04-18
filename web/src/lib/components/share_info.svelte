<script lang="ts">
    import type { List } from "$lib/models/list";
    import type { Trail } from "$lib/models/trail";
    import type { UserAnonymous } from "$lib/models/user";
    import { currentUser, users_show } from "$lib/stores/user_store";
    import { getFileURL } from "$lib/util/file_util";
    import { _ } from "svelte-i18n";
    import { fly } from "svelte/transition";

    interface Props {
        subject: Trail | List;
        large?: boolean;
        type: "trail" | "list";
    }

    let { subject, large = false, type }: Props = $props();

    const shareData =
        type == "trail"
            ? (subject as Trail).expand?.trail_share_via_trail
            : (subject as List).expand?.list_share_via_list;

    let showInfo: boolean = $state(false);
    let loading: boolean = $state(false);
    let infoLoaded: boolean = $state(false);
    let subjectIsOwned: boolean = subject.author == $currentUser?.id;
    let author: UserAnonymous | undefined = $state();

    async function fetchInfo() {
        if (!infoLoaded) {
            loading = true;
            loading = false;
            if (subjectIsOwned) {
                for (const share of shareData ?? []) {
                    share.expand = {
                        user: await users_show(share.user),
                    };
                }
            } else {
                author = await users_show(subject.author!);
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
        onmouseenter={fetchInfo}
        onmouseleave={() => (showInfo = false)}
    ></i>
    {#if showInfo}
        <ul
            class="menu absolute z-10 top-8 bg-menu-background border border-input-border rounded-xl shadow-md overflow-hidden p-3 space-y-3 text-content -translate-x-3/4"
            in:fly={{ y: -10, duration: 150 }}
            out:fly={{ y: -10, duration: 150 }}
        >
            {#if loading}
                <div class="spinner spinner-dark"></div>
            {:else if infoLoaded && shareData}
                {#if subjectIsOwned}
                    {#each shareData as share}
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
                            {#if shareData[0].permission == "view"}
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
